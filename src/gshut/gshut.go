package gshut

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Component описывает абстрактный ресурс с двумя режимами остановки.
type Component interface {
	// GracefulStop плавно завершает работу (ждет текущие запросы, сбрасывает буферы).
	GracefulStop(ctx context.Context) error
	// Stop принудительно обрывает работу, если GracefulStop не уложился в таймаут.
	Stop() error
}

// ComponentFunc позволяет оборачивать обычные функции в интерфейс Component.
type ComponentFunc struct {
	OnGraceful func(ctx context.Context) error
	OnStop     func() error
}

func (f ComponentFunc) GracefulStop(ctx context.Context) error {
	if f.OnGraceful != nil {
		return f.OnGraceful(ctx)
	}
	return nil
}

func (f ComponentFunc) Stop() error {
	if f.OnStop != nil {
		return f.OnStop()
	}
	return nil
}

type node struct {
	name         string
	component    Component
	timeout      time.Duration
	dependents   []string      // Кто зависит от этого узла (должны остановиться РАНЬШЕ)
	dependencies []string      // От кого зависит этот узел (остановятся ПОЗЖЕ)
	doneCh       chan struct{} // Закрывается, когда этот узел полностью остановлен
}

// Manager управляет графом остановки.
type Manager struct {
	mu    sync.Mutex
	nodes map[string]*node
}

func NewManager() *Manager {
	return &Manager{
		nodes: make(map[string]*node),
	}
}

// Register регистрирует компонент в графе.
// name — уникальное имя компонента (например, "db", "grpc").
// timeout — время, выделенное конкретно на GracefulStop этого узла.
// dependsOn — имена компонентов, которые должны остаться живыми, пока этот узел не остановится.
func (m *Manager) Register(name string, comp Component, timeout time.Duration, dependsOn ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.nodes[name]; exists {
		return fmt.Errorf("компонент %q уже зарегистрирован", name)
	}

	m.nodes[name] = &node{
		name:         name,
		component:    comp,
		timeout:      timeout,
		dependencies: dependsOn,
		doneCh:       make(chan struct{}),
	}
	return nil
}

// Shutdown запускает остановку всего графа.
func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	if err := m.buildAndValidateGraph(); err != nil {
		m.mu.Unlock()
		return err
	}
	nodesSnapshot := make([]*node, 0, len(m.nodes))
	for _, n := range m.nodes {
		nodesSnapshot = append(nodesSnapshot, n)
	}
	m.mu.Unlock()

	var (
		wg   sync.WaitGroup
		errs []error
		errM sync.Mutex
	)

	recordErr := func(err error) {
		if err != nil {
			errM.Lock()
			errs = append(errs, err)
			errM.Unlock()
		}
	}

	// Запускаем горутину для каждого узла. Они сами синхронизируются через каналы.
	for _, n := range nodesSnapshot {
		wg.Add(1)
		go func(curr *node) {
			defer wg.Done()
			defer close(curr.doneCh) // Сигнализируем тем, от кого мы зависим, что мы остановлены

			// 1. Ждем остановки всех компонентов, которые зависят от текущего узла
			for _, depName := range curr.dependents {
				depNode := m.nodes[depName]
				select {
				case <-depNode.doneCh:
				case <-ctx.Done():
					// Если общий контекст уже истек, не ждем и переходим к принудительной остановке
				}
			}

			// 2. Пытаемся выполнить GracefulStop с индивидуальным таймаутом узла
			nodeCtx, cancel := context.WithTimeout(ctx, curr.timeout)
			defer cancel()

			gracefulDone := make(chan error, 1)
			go func() {
				gracefulDone <- curr.component.GracefulStop(nodeCtx)
			}()

			select {
			case err := <-gracefulDone:
				if err != nil {
					recordErr(fmt.Errorf("[%s] ошибка graceful stop: %w", curr.name, err))
					// Если плавная остановка упала с ошибкой, пробуем жесткий Stop
					if stopErr := curr.component.Stop(); stopErr != nil {
						recordErr(fmt.Errorf("[%s] ошибка fallback stop: %w", curr.name, stopErr))
					}
				}
			case <-nodeCtx.Done():
				// Таймаут узла или общий контекст истек — вызываем жесткий Stop
				recordErr(fmt.Errorf("[%s] таймаут graceful stop (%v), вызываем принудительный Stop", curr.name, curr.timeout))
				if stopErr := curr.component.Stop(); stopErr != nil {
					recordErr(fmt.Errorf("[%s] ошибка принудительного stop: %w", curr.name, stopErr))
				}
			}
		}(n)
	}

	wg.Wait()
	return errors.Join(errs...)
}

// buildAndValidateGraph строит обратные связи и проверяет граф на циклы (DFS).
func (m *Manager) buildAndValidateGraph() error {
	for _, n := range m.nodes {
		n.dependents = nil
	}

	for name, n := range m.nodes {
		for _, depName := range n.dependencies {
			target, exists := m.nodes[depName]
			if !exists {
				return fmt.Errorf("компонент %q зависит от несуществующего компонента %q", name, depName)
			}
			target.dependents = append(target.dependents, name)
		}
	}

	// Проверка на циклические зависимости (0 = не посещен, 1 = в процессе, 2 = проверен)
	state := make(map[string]int)
	var visit func(name string) error
	visit = func(name string) error {
		state[name] = 1
		for _, depName := range m.nodes[name].dependencies {
			if state[depName] == 1 {
				return fmt.Errorf("обнаружена циклическая зависимость с участием %q и %q", name, depName)
			}
			if state[depName] == 0 {
				if err := visit(depName); err != nil {
					return err
				}
			}
		}
		state[name] = 2
		return nil
	}

	for name := range m.nodes {
		if state[name] == 0 {
			if err := visit(name); err != nil {
				return err
			}
		}
	}
	return nil
}
