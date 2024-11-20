package gocodevec

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// 项目信息
type ProjectInfo struct {
	Packages map[string]*PackageInfo // key: 包路径
}

// 包信息
type PackageInfo struct {
	Name       string
	Path       string
	Files      map[string]*FileInfo // key: 文件路径
	Imports    []string
	Interfaces map[string]*InterfaceInfo // key: 接口名
	Structs    map[string]*StructInfo    // key: 结构体名
	Functions  map[string]*FunctionInfo  // key: 函数名
}

// 文件信息
type FileInfo struct {
	Path       string
	Package    string
	Imports    []string
	Doc        string
	Functions  map[string]*FunctionInfo  // 添加这些字段
	Structs    map[string]*StructInfo    // 添加这些字段
	Interfaces map[string]*InterfaceInfo // 添加这些字段
}

// 接口信息
type InterfaceInfo struct {
	Name    string
	Methods []*MethodInfo
	Doc     string
}

// 结构体信息
type StructInfo struct {
	Name       string
	Fields     []*FieldInfo
	Methods    []*MethodInfo
	Implements []string // 实现的接口列表
	Doc        string
}

// 函数信息
type FunctionInfo struct {
	Name       string
	Params     []*ParamInfo
	Results    []*ParamInfo
	Doc        string
	IsExported bool
	Body       string
}

type FieldInfo struct {
	Name string
	Type string
	Tag  string
	Doc  string
}

// 定义方法信息
type MethodInfo struct {
	Name     string
	Receiver string
	Params   []*ParamInfo
	Results  []*ParamInfo
	Doc      string
	Body     string
}

// 定义参数信息
type ParamInfo struct {
	Name string
	Type string
}

func ParseProject(projectPath string) (*ProjectInfo, error) {
	projectInfo := &ProjectInfo{
		Packages: make(map[string]*PackageInfo),
	}

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			// 获取相对于项目根目录的包路径
			relPath, err := filepath.Rel(projectPath, filepath.Dir(path))
			if err != nil {
				return err
			}

			// 解析文件
			fileInfo, err := parseFile(path)
			if err != nil {
				return err
			}

			// 使用相对路径作为包的唯一标识
			if _, exists := projectInfo.Packages[relPath]; !exists {
				projectInfo.Packages[relPath] = &PackageInfo{
					Name:       fileInfo.Package,
					Path:       relPath,
					Files:      make(map[string]*FileInfo),
					Structs:    make(map[string]*StructInfo),
					Interfaces: make(map[string]*InterfaceInfo),
					Functions:  make(map[string]*FunctionInfo),
				}
			}

			pkg := projectInfo.Packages[relPath]
			pkg.Files[path] = fileInfo
			for name, fn := range fileInfo.Functions {
				pkg.Functions[name] = fn
			}
			for name, st := range fileInfo.Structs {
				pkg.Structs[name] = st
			}
			for name, iface := range fileInfo.Interfaces {
				pkg.Interfaces[name] = iface
			}
		}
		return nil
	})

	return projectInfo, err
}

func parseFile(filePath string) (*FileInfo, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	fileInfo := &FileInfo{
		Path:       filePath,
		Package:    node.Name.Name,
		Functions:  make(map[string]*FunctionInfo),
		Structs:    make(map[string]*StructInfo),
		Interfaces: make(map[string]*InterfaceInfo),
	}

	// 收集导入信息
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, "\"")
		fileInfo.Imports = append(fileInfo.Imports, importPath)
	}

	// 遍历AST
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			// 处理接口定义
			if interfaceType, ok := x.Type.(*ast.InterfaceType); ok {
				iface := parseInterface(x, interfaceType)
				fileInfo.Interfaces[iface.Name] = iface
			}

			// 处理结构体定义
			if structType, ok := x.Type.(*ast.StructType); ok {
				st := parseStruct(x, structType)
				fileInfo.Structs[st.Name] = st
			}

		case *ast.FuncDecl:
			// 处理方法
			if x.Recv != nil {
				method := parseMethod(x)
				if receiverType := getReceiverType(x.Recv); receiverType != "" {
					if st, ok := fileInfo.Structs[receiverType]; ok {
						st.Methods = append(st.Methods, method)
					}
				}
			} else {
				// 处理普通函数
				fn := parseFunction(x)
				fileInfo.Functions[fn.Name] = fn
			}
		}
		return true
	})

	// 分析接口实现关系
	analyzeInterfaceImplementations(fileInfo)

	return fileInfo, nil
}

func analyzeInterfaceImplementations(file *FileInfo) {
	for _, st := range file.Structs {
		for _, iface := range file.Interfaces {
			if implementsInterface(st, iface) {
				st.Implements = append(st.Implements, iface.Name)
			}
		}
	}
}

// 检查结构体是否实现了接口
func implementsInterface(st *StructInfo, iface *InterfaceInfo) bool {
	for _, requiredMethod := range iface.Methods {
		found := false
		for _, method := range st.Methods {
			if methodsMatch(method, requiredMethod) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// 辅助函数检查方法签名是否匹配
func methodsMatch(m1, m2 *MethodInfo) bool {
	if m1.Name != m2.Name {
		return false
	}
	// 检查参数
	if !paramsMatch(m1.Params, m2.Params) {
		return false
	}
	// 检查返回值
	return paramsMatch(m1.Results, m2.Results)
}

// 解析接口定义
func parseInterface(typeSpec *ast.TypeSpec, interfaceType *ast.InterfaceType) *InterfaceInfo {
	iface := &InterfaceInfo{
		Name:    typeSpec.Name.Name,
		Methods: make([]*MethodInfo, 0),
	}

	// 获取接口文档
	if typeSpec.Doc != nil {
		iface.Doc = typeSpec.Doc.Text()
	}

	// 解析接口方法
	for _, method := range interfaceType.Methods.List {
		if methodType, ok := method.Type.(*ast.FuncType); ok {
			for _, name := range method.Names {
				methodInfo := &MethodInfo{
					Name:    name.Name,
					Params:  parseParams(methodType.Params),
					Results: parseParams(methodType.Results),
				}
				if method.Doc != nil {
					methodInfo.Doc = method.Doc.Text()
				}
				iface.Methods = append(iface.Methods, methodInfo)
			}
		}
	}

	return iface
}

// 解析结构体定义
func parseStruct(typeSpec *ast.TypeSpec, structType *ast.StructType) *StructInfo {
	st := &StructInfo{
		Name:       typeSpec.Name.Name,
		Fields:     make([]*FieldInfo, 0),
		Methods:    make([]*MethodInfo, 0),
		Implements: make([]string, 0),
	}

	// 获取结构体文档
	if typeSpec.Doc != nil {
		st.Doc = typeSpec.Doc.Text()
	}

	// 解析结构体字段
	for _, field := range structType.Fields.List {
		fieldInfo := parseField(field)
		st.Fields = append(st.Fields, fieldInfo)
	}

	return st
}

// 解析字段
func parseField(field *ast.Field) *FieldInfo {
	fieldInfo := &FieldInfo{
		Type: formatExpr(field.Type),
	}

	// 获取字段文档
	if field.Doc != nil {
		fieldInfo.Doc = field.Doc.Text()
	}

	// 获取字段标签
	if field.Tag != nil {
		fieldInfo.Tag = field.Tag.Value
	}

	// 处理字段名称
	if len(field.Names) > 0 {
		fieldInfo.Name = field.Names[0].Name
	} else {
		// 匿名字段，使用类型作为名
		fieldInfo.Name = fieldInfo.Type
	}

	return fieldInfo
}

// 解析方法
func parseMethod(funcDecl *ast.FuncDecl) *MethodInfo {
	method := &MethodInfo{
		Name:     funcDecl.Name.Name,
		Params:   parseParams(funcDecl.Type.Params),
		Results:  parseParams(funcDecl.Type.Results),
		Receiver: getReceiverType(funcDecl.Recv),
	}

	// 获取方法文档
	if funcDecl.Doc != nil {
		method.Doc = funcDecl.Doc.Text()
	}

	// 添加方法体解析
	if funcDecl.Body != nil {
		fset := token.NewFileSet()
		var buf bytes.Buffer
		printer.Fprint(&buf, fset, funcDecl.Body)
		method.Body = buf.String()
	}

	return method
}

// 解析函数
func parseFunction(funcDecl *ast.FuncDecl) *FunctionInfo {
	fn := &FunctionInfo{
		Name:       funcDecl.Name.Name,
		Params:     parseParams(funcDecl.Type.Params),
		Results:    parseParams(funcDecl.Type.Results),
		IsExported: ast.IsExported(funcDecl.Name.Name),
	}

	// 获取函数文档
	if funcDecl.Doc != nil {
		fn.Doc = funcDecl.Doc.Text()
	}

	// 添加函数体解析
	if funcDecl.Body != nil {
		// 获取函数体的源代码
		fset := token.NewFileSet()
		var buf bytes.Buffer
		printer.Fprint(&buf, fset, funcDecl.Body)
		fn.Body = buf.String()
	}

	return fn
}

// 获取接收者类型
func getReceiverType(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}

	switch t := recv.List[0].Type.(type) {
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	case *ast.Ident:
		return t.Name
	}
	return ""
}

// 格式化类型表达式
func formatExpr(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + formatExpr(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + formatExpr(t.Elt)
		}
		return fmt.Sprintf("[%s]%s", formatExpr(t.Len), formatExpr(t.Elt))
	case *ast.SelectorExpr:
		return formatExpr(t.X) + "." + t.Sel.Name
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", formatExpr(t.Key), formatExpr(t.Value))
	case *ast.ChanType:
		switch t.Dir {
		case ast.SEND:
			return fmt.Sprintf("chan<- %s", formatExpr(t.Value))
		case ast.RECV:
			return fmt.Sprintf("<-chan %s", formatExpr(t.Value))
		default:
			return fmt.Sprintf("chan %s", formatExpr(t.Value))
		}
	case *ast.FuncType:
		return "func" // 可以进一步展开函数类型的详细信息
	case *ast.StructType:
		return "struct" // 可以进一步展开结构体类型的详细信息
	case *ast.BasicLit:
		return t.Value
	default:
		return fmt.Sprintf("<%T>", expr)
	}
}

// 检查参数列表是否匹配
func paramsMatch(p1, p2 []*ParamInfo) bool {
	if len(p1) != len(p2) {
		return false
	}
	for i := range p1 {
		if p1[i].Type != p2[i].Type {
			return false
		}
	}
	return true
}

// 解析参数列表
func parseParams(fields *ast.FieldList) []*ParamInfo {
	if fields == nil {
		return nil
	}

	var params []*ParamInfo
	for _, field := range fields.List {
		paramType := formatExpr(field.Type)

		// 如果参数有名称
		if len(field.Names) > 0 {
			for _, name := range field.Names {
				params = append(params, &ParamInfo{
					Name: name.Name,
					Type: paramType,
				})
			}
		} else {
			// 匿名参数（通常是返回值）
			params = append(params, &ParamInfo{
				Name: "",
				Type: paramType,
			})
		}
	}

	return params
}
