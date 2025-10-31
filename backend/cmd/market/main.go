package main

import (
	"fmt"
	"os"

	"github.com/kymo-mcp/mcpcan/internal/market/app"
)

func main() {
	// 创建应用程序实例
	appInstance, err := app.New()
	if err != nil {
		fmt.Printf("创建应用程序实例失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化应用程序
	if err := appInstance.Initialize(); err != nil {
		fmt.Printf("初始化应用程序失败: %v\n", err)
		os.Exit(1)
	}

	// 运行应用程序
	if err := appInstance.Run(); err != nil {
		fmt.Printf("运行应用程序失败: %v\n", err)
		os.Exit(1)
	}
}
