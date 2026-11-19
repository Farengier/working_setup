# Использование Kubernetes с Colima
colima kubernetes

## Создание инфраструктуры

### PostgreSQL
```bash
kubectl apply -f k8s/postgres/pvc.yaml
kubectl apply -f k8s/postgres/deployment.yaml
kubectl apply -f k8s/postgres/service.yaml
```

### Temporal
```bash
kubectl apply -f k8s/temporal/deployment.yaml
kubectl apply -f k8s/temporal/service.yaml
```

### Temporal UI
```bash
kubectl apply -f k8s/temporal/ui/deployment.yaml
kubectl apply -f k8s/temporal/ui/service.yaml
```

## Проверка ресурсов
```bash
kubectl get pods
kubectl get services
kubectl get deployments
```

## Доступ к системам

### PostgreSQL
```bash
kubectl exec -it deployment/postgres -- psql -U postgres -d agent_space
```

### Temporal gRPC (для приложений внутри кластера)
Используйте сервисное имя: `temporal-svc:7233`

### Temporal UI (для доступа из браузера)
```bash
kubectl port-forward svc/temporal-ui-svc 8080:8080
```
Затем откройте http://localhost:8080 в браузере

### Temporal gRPC (для локального подключения через port-forward)
```bash
kubectl port-forward svc/temporal-svc 7233:7233
```

## Разработка через mirrord

Для разработки локального приложения с доступом к инфраструктуре в Kubernetes:

```bash
# Установка mirrord
brew install metalbear-co/mirrord/mirrord

# Запуск приложения с доступом к кластеру
mirrord exec --steal=false -- go run cmd/ai-assistant-ahp-memory-manager/main.go
```

### Переменные окружения для локальной разработки

Для подключения к Temporal в Kubernetes добавьте в `.env`:
```
TEMPORAL_ADDRESS=temporal-svc:7233
```

или для port-forward:
```
TEMPORAL_ADDRESS=localhost:7233
```

## Структура манифестов

```
k8s/
├── postgres/
│   ├── deployment.yaml
│   ├── pvc.yaml
│   └── service.yaml
└── temporal/
    ├── deployment.yaml
    ├── service.yaml
    └── ui/
        ├── deployment.yaml
        └── service.yaml
```
