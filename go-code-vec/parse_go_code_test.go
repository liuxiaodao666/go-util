package gocodevec

import (
	"fmt"
	"log"
	"testing"
)

func Test_parseGoFile(t *testing.T) {
	parseGoFile()
}

func Test_parseprojject(t *testing.T) {
	projectInfo, err := ParseProject(`D:\\code\\go\\go-util`)
	if err != nil {
		log.Fatal(err)
	}

	// 遍历并打印项目信息
	for pkgPath, pkg := range projectInfo.Packages {
		fmt.Printf("包: %s (%s)\n", pkg.Name, pkgPath)

		// 打印结构体信息
		for _, st := range pkg.Structs {
			fmt.Printf("  结构体: %s\n", st.Name)
			fmt.Printf("    实现接口: %v\n", st.Implements)
			if len(st.Methods) > 0 {
				fmt.Printf("    方法: %v\n", st.Methods[0].Body)
			}

		}

		// 打印接口信息
		for _, iface := range pkg.Interfaces {
			fmt.Printf("  接口: %s\n", iface.Name)
			fmt.Printf("    方法数量: %d\n", len(iface.Methods))
		}

		for _, functions := range pkg.Functions {
			fmt.Printf("  函数名: %s\n", functions.Name)
			fmt.Printf("    函数参数: %v\n", len(functions.Params))
			fmt.Printf("    函数Results: %v\n", len(functions.Results))
			fmt.Printf("    函数body: %v\n", functions.Body)
		}
	}
}
