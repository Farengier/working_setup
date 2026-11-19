package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode"

	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	var targetDir string
	opts := protogen.Options{
		ParamFunc: func(name, value string) error {
			if name == "target_dir" {
				targetDir = value
			}
			return nil
		},
	}

	opts.Run(func(gen *protogen.Plugin) error {
		if targetDir == "" {
			targetDir = "internal/controller"
		}

		for _, f := range gen.Files {
			if !f.Generate || len(f.Services) == 0 {
				continue
			}
			generateHandlers(gen, f, targetDir)
		}
		return nil
	})
}

func generateHandlers(gen *protogen.Plugin, file *protogen.File, targetDir string) {
	for _, service := range file.Services {
		pkgName := toSnakeCase(service.GoName)
		serviceDir := path.Join(targetDir, pkgName)
		serviceImportPath := getTargetImportPath(serviceDir)
		srvFileName := path.Join(serviceDir, "service.go")

		fmt.Fprintln(os.Stderr, serviceDir)
		fmt.Fprintln(os.Stderr, serviceImportPath)
		fmt.Fprintln(os.Stderr, srvFileName)
		if _, err := os.Stat(srvFileName); err != nil {
			writeBaseServiceFile(gen, srvFileName, serviceImportPath, pkgName, service, file)
		}

		for _, method := range service.Methods {
			methodFileName := path.Join(serviceDir, toSnakeCase(method.GoName)+".go")

			// Проверяем существование именно файла метода, а не service.go
			if _, err := os.Stat(methodFileName); err == nil {
				continue
			}

			g := gen.NewGeneratedFile(methodFileName, serviceImportPath)
			g.P("package ", pkgName)
			g.P()

			// QualifiedGoIdent автоматически добавит блок import (...) со всеми нужными пакетами
			ctxType := g.QualifiedGoIdent(protogen.GoIdent{
				GoName:       "Context",
				GoImportPath: "context",
			})
			codesUnimplemented := g.QualifiedGoIdent(protogen.GoIdent{
				GoName:       "Unimplemented",
				GoImportPath: "google.golang.org/grpc/codes",
			})
			statusErrorf := g.QualifiedGoIdent(protogen.GoIdent{
				GoName:       "Errorf",
				GoImportPath: "google.golang.org/grpc/status",
			})

			// Автоматически подтянет импорт contracts и отформатирует вложенные структуры
			reqType := g.QualifiedGoIdent(method.Input.GoIdent)
			respType := g.QualifiedGoIdent(method.Output.GoIdent)

			g.P("func (s *", service.GoName, ") ", method.GoName, "(ctx ", ctxType, ", req *", reqType, ") (*", respType, ", error) {")
			g.P("\treturn nil, ", statusErrorf, "(", codesUnimplemented, ", \"method ", method.GoName, " not implemented\")")
			g.P("}")
		}
	}
}

func writeBaseServiceFile(
	gen *protogen.Plugin,
	srvFileName string,
	serviceImportPath protogen.GoImportPath,
	pkgName string,
	service *protogen.Service,
	file *protogen.File,
) {
	srvFile := gen.NewGeneratedFile(srvFileName, serviceImportPath)
	srvFile.P("package ", pkgName)
	srvFile.P()

	unimplementedServer := srvFile.QualifiedGoIdent(protogen.GoIdent{
		GoName:       "Unimplemented" + service.GoName + "Server",
		GoImportPath: file.GoImportPath,
	})

	srvFile.P("type ", service.GoName, " struct {")
	srvFile.P("\t", unimplementedServer)
	srvFile.P("}")
}

func toSnakeCase(s string) string {
	runes := []rune(s)
	var res []rune

	for i := 0; i < len(runes); i++ {
		if i > 0 && unicode.IsUpper(runes[i]) {
			// Предыдущая буква строчная (например, memory|Manager)
			prevLower := unicode.IsLower(runes[i-1])
			// Следующая буква строчная — это конец аббревиатуры (например, AH|PService -> P + s)
			nextLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])

			if prevLower || nextLower {
				res = append(res, '_')
			}
		}
		res = append(res, unicode.ToLower(runes[i]))
	}
	return string(res)
}

// getTargetImportPath находит go.mod и склеивает имя модуля с относительным путем (например, "internal/service")
func getTargetImportPath(relDir string) protogen.GoImportPath {
	dir, err := os.Getwd()
	if err != nil {
		return protogen.GoImportPath(relDir)
	}

	for {
		modFile := filepath.Join(dir, "go.mod")
		if data, err := os.ReadFile(modFile); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "module ") {
					modName := strings.TrimSpace(strings.TrimPrefix(line, "module"))
					// Используем path.Join (прямые слеши), так как это Go import path
					return protogen.GoImportPath(path.Join(modName, filepath.ToSlash(relDir)))
				}
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // дошли до корня диска
		}
		dir = parent
	}

	return protogen.GoImportPath(relDir)
}
