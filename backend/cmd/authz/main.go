package main

import (
	"log"
	"os"

	"qm-mcp-server/internal/authz/app"
)

// main 主函数
func main() {
	// 创建应用程序实例
	appInstance := app.New()

	// 初始化应用程序
	if err := appInstance.Initialize(); err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
		os.Exit(1)
	}

	// 运行应用程序
	if err := appInstance.Run(); err != nil {
		log.Fatalf("Failed to run application: %v", err)
		os.Exit(1)
	}
}
