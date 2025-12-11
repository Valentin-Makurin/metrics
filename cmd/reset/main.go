package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func main() {
	// Находим директорию, где находится утилита
	exePath, err := os.Executable()
	if err != nil {
		exePath, _ = os.Getwd()
	}
	utilDir := filepath.Dir(exePath)

	rootDir := filepath.Dir(filepath.Dir(utilDir))

	fmt.Printf("Scanning from root directory: %s\n", rootDir)

	// Находим абсолютный путь к корневой директории
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting absolute path: %v\n", err)
		os.Exit(1)
	}

	// Мапа для хранения сгенерированных методов по пакетам
	packages := make(map[string][]generatedMethod)

	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Проверяем, не является ли это директорией самой утилиты генератора
		if d.IsDir() {

			absPath, _ := filepath.Abs(path)
			if absPath == utilDir {
				return filepath.SkipDir
			}

			base := filepath.Base(absPath)
			if strings.HasPrefix(base, ".") || base == "vendor" || base == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		if strings.HasSuffix(path, "reset.gen.go") {
			return nil
		}

		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		absPath, _ := filepath.Abs(path)
		if strings.HasPrefix(absPath, utilDir) {
			return nil
		}

		methods, err := processFile(path)
		if err != nil {
			return fmt.Errorf("processing file %s: %w", path, err)
		}

		if len(methods) > 0 {
			pkgDir := filepath.Dir(path)
			packages[pkgDir] = append(packages[pkgDir], methods...)
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
		os.Exit(1)
	}

	// Генерируем файлы для каждого пакета
	generatedCount := 0
	for pkgDir, methods := range packages {
		if err := generateResetFile(pkgDir, methods); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating file for %s: %v\n", pkgDir, err)
			continue
		}
		generatedCount++
		fmt.Printf("Generated reset.gen.go for package: %s\n", pkgDir)
	}

	if generatedCount > 0 {
		fmt.Printf("\nSuccessfully generated %d reset.gen.go files!\n", generatedCount)
	} else {
		fmt.Println("No structures with // generate:reset comment found.")
	}
}

type generatedMethod struct {
	StructName string
	Fields     []fieldInfo
}

type fieldInfo struct {
	Name     string
	Type     string
	IsPtr    bool
	IsSlice  bool
	IsMap    bool
	IsStruct bool
}

func processFile(filename string) ([]generatedMethod, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var methods []generatedMethod

	// Проходим по всем декларациям в файле
	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		if genDecl.Doc == nil {
			continue
		}

		// Проверяем комментарии для структуры
		hasResetComment := false
		for _, comment := range genDecl.Doc.List {
			if strings.Contains(comment.Text, "generate:reset") {
				hasResetComment = true
				break
			}
		}

		if !hasResetComment {
			continue
		}

		// Обрабатываем каждую спецификацию
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			structName := typeSpec.Name.Name
			fields := extractFields(structType)

			methods = append(methods, generatedMethod{
				StructName: structName,
				Fields:     fields,
			})
		}
	}

	return methods, nil
}

func extractFields(structType *ast.StructType) []fieldInfo {
	var fields []fieldInfo

	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {

			if ident, ok := field.Type.(*ast.Ident); ok && ident != nil {
				fieldName := ident.Name
				if !unicode.IsUpper(rune(fieldName[0])) {
					continue // Пропускаем неэкспортируемые анонимные поля
				}

				fieldType := getTypeName(field.Type)
				isPtr := isPointerType(field.Type)
				isSlice := isSliceType(field.Type)
				isMap := isMapType(field.Type)
				isStruct := isStructType(field.Type)

				fields = append(fields, fieldInfo{
					Name:     fieldName,
					Type:     fieldType,
					IsPtr:    isPtr,
					IsSlice:  isSlice,
					IsMap:    isMap,
					IsStruct: isStruct,
				})
			}
			continue
		}

		fieldName := field.Names[0].Name
		if !unicode.IsUpper(rune(fieldName[0])) {
			continue // Пропускаем неэкспортируемые поля
		}

		fieldType := getTypeName(field.Type)
		isPtr := isPointerType(field.Type)
		isSlice := isSliceType(field.Type)
		isMap := isMapType(field.Type)
		isStruct := isStructType(field.Type)

		fields = append(fields, fieldInfo{
			Name:     fieldName,
			Type:     fieldType,
			IsPtr:    isPtr,
			IsSlice:  isSlice,
			IsMap:    isMap,
			IsStruct: isStruct,
		})
	}

	return fields
}

func getTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + getTypeName(t.X)
	case *ast.ArrayType:
		return "[]" + getTypeName(t.Elt)
	case *ast.MapType:
		key := getTypeName(t.Key)
		value := getTypeName(t.Value)
		return fmt.Sprintf("map[%s]%s", key, value)
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", getTypeName(t.X), t.Sel.Name)
	default:
		return fmt.Sprintf("%v", t)
	}
}

func isPointerType(expr ast.Expr) bool {
	_, ok := expr.(*ast.StarExpr)
	return ok
}

func isSliceType(expr ast.Expr) bool {
	arrType, ok := expr.(*ast.ArrayType)
	if !ok {
		return false
	}
	// Проверяем, что это именно слайс (без указания размера)
	return arrType.Len == nil
}

func isMapType(expr ast.Expr) bool {
	_, ok := expr.(*ast.MapType)
	return ok
}

func isStructType(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.Ident:
		// Проверяем, начинается ли тип с заглавной буквы (экспортируемый)
		return len(t.Name) > 0 && unicode.IsUpper(rune(t.Name[0]))
	case *ast.StarExpr:
		return isStructType(t.X)
	case *ast.SelectorExpr:
		return true
	case *ast.ArrayType:
		return false
	case *ast.MapType:
		return false
	default:
		return false
	}
}

func generateResetFile(pkgDir string, methods []generatedMethod) error {
	if len(methods) == 0 {
		return nil
	}

	pkgName, err := getPackageName(pkgDir)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	// Заголовок
	buf.WriteString(fmt.Sprintf("// Code generated by reset tool; DO NOT EDIT.\n\n"))
	buf.WriteString(fmt.Sprintf("package %s\n\n", pkgName))

	hasTimeImport := false
	for _, method := range methods {
		for _, field := range method.Fields {
			if strings.Contains(field.Type, "time.") {
				hasTimeImport = true
				break
			}
		}
		if hasTimeImport {
			break
		}
	}

	if hasTimeImport {
		buf.WriteString("import \"time\"\n\n")
	}

	// Генерируем методы для каждой структуры
	for _, method := range methods {
		buf.WriteString(generateMethod(method))
		buf.WriteString("\n\n")
	}

	// Форматируем код
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("formatting source: %w", err)
	}

	// Записываем в файл
	outputPath := filepath.Join(pkgDir, "reset.gen.go")
	return os.WriteFile(outputPath, formatted, 0644)
}

func getPackageName(dir string) (string, error) {
	// Ищем любой .go файл в директории
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".go") &&
			!strings.HasSuffix(entry.Name(), "_test.go") &&
			!strings.HasSuffix(entry.Name(), "reset.gen.go") {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, filepath.Join(dir, entry.Name()), nil, parser.PackageClauseOnly)
			if err != nil {
				continue
			}
			return node.Name.Name, nil
		}
	}

	return "", fmt.Errorf("no .go files found in directory %s", dir)
}

func generateMethod(method generatedMethod) string {
	var buf bytes.Buffer

	receiver := strings.ToLower(method.StructName[:1])
	if receiver == "s" || receiver == "r" {
		receiver = "x"
	}

	buf.WriteString(fmt.Sprintf("func (%s *%s) Reset() {\n", receiver, method.StructName))
	buf.WriteString(fmt.Sprintf("\tif %s == nil {\n", receiver))
	buf.WriteString("\t\treturn\n")
	buf.WriteString("\t}\n\n")

	// Генерируем код
	for _, field := range method.Fields {
		buf.WriteString(generateFieldReset(receiver, field))
	}

	buf.WriteString("}")

	return buf.String()
}

func generateFieldReset(receiver string, field fieldInfo) string {
	accessor := fmt.Sprintf("%s.%s", receiver, field.Name)

	switch {
	case field.IsMap:
		return fmt.Sprintf("\tclear(%s)\n", accessor)

	case field.IsSlice:
		return fmt.Sprintf("\t%s = %s[:0]\n", accessor, accessor)

	case field.IsPtr:
		// Для указателей нужно проверить на nil и сбросить значение
		if field.IsStruct && !strings.HasPrefix(field.Type, "[]") && !strings.HasPrefix(field.Type, "map[") {
			// Указатель на структуру, которая может иметь метод Reset
			return fmt.Sprintf("\tif %s != nil {\n\t\tif resetter, ok := interface{}(%s).(interface{ Reset() }); ok {\n\t\t\tresetter.Reset()\n\t\t}\n\t}\n",
				accessor, accessor)
		} else {
			// Указатель на примитив или другую неструктуру
			return fmt.Sprintf("\tif %s != nil {\n\t\t*%s = %s\n\t}\n",
				accessor, accessor, getZeroValue(field.Type))
		}

	case field.IsStruct:
		// Поле структуры (не указатель)
		return fmt.Sprintf("\tif resetter, ok := interface{}(&%s).(interface{ Reset() }); ok {\n\t\tresetter.Reset()\n\t}\n",
			accessor)

	default:
		// Примитивные типы
		return fmt.Sprintf("\t%s = %s\n", accessor, getZeroValue(field.Type))
	}
}

func getZeroValue(typeName string) string {
	// Убираем указатели для определения базового типа
	baseType := strings.TrimPrefix(typeName, "*")

	switch baseType {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"byte", "rune":
		return "0"
	case "float32", "float64":
		return "0.0"
	case "complex64", "complex128":
		return "0+0i"
	case "bool":
		return "false"
	case "string":
		return `""`
	case "time.Time":
		return "time.Time{}"
	default:
		// Для неизвестных типов возвращаем нулевое значение через var
		return fmt.Sprintf("%s{}", baseType)
	}
}
