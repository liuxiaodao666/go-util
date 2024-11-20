package gocodevec

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"

	"log"
)

func parseGoFile() {
	fset := token.NewFileSet() // 创建一个新的FileSet
	filePath := "D:\\code\\go\\go-util\\logger\\logger.go"
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	// 读取文件内容
	src, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	// 使用bufio.Scanner来按行读取文件内容
	scanner := bufio.NewScanner(bytes.NewReader(src))

	// 遍历AST节点
	ast.Inspect(file, func(n ast.Node) bool {
		// 检查节点是否是函数声明
		fn, ok := n.(*ast.FuncDecl)
		if ok {
			// 打印函数名称
			fmt.Printf("Function: %s\n", fn.Name.Name)
			// 获取函数体的起始和结束位置
			start := fset.Position(fn.Body.Lbrace).Offset
			end := fset.Position(fn.Body.Rbrace).Offset
			// 提取函数体内容
			body := src[start : end+1]
			// 打印函数体
			fmt.Printf("Body:\n%s\n", string(body))
			fmt.Println() // 打印一个空行作为分隔
		}
		return true
	})

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
