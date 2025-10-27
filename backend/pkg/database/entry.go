package database

import (
	"time"

	"qm-mcp-server/pkg/common"
	"qm-mcp-server/pkg/database/repository/mysql"
)

// Init 初始化数据库连接
func Init(databaseConfig *common.DatabaseConfig) error {
	// 初始化 MySQL 配置
	mysqlConfig := &mysql.Config{
		Host:                databaseConfig.MySQL.Host,
		Port:                databaseConfig.MySQL.Port,
		Username:            databaseConfig.MySQL.Username,
		Password:            databaseConfig.MySQL.Password,
		Database:            databaseConfig.MySQL.Database,
		ConnectTimeout:      60 * time.Second,
		MaxIdleConns:        10,
		MaxOpenConns:        100,
		HealthCheckInterval: 30 * time.Second,
		MaxRetries:          3,
		RetryInterval:       5 * time.Second,
	}
	err := mysql.InitDB(mysqlConfig)
	if err != nil {
		return err
	}

	// 初始化数据库连接、

	return nil
}
