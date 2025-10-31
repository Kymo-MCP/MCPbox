package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/common"

	"github.com/redis/go-redis/v9"
)

// Client Redis客户端结构
type Client struct {
	config *common.RedisConfig
	client *redis.Client
}

var globalClient *Client

// Init 初始化Redis客户端
func Init(config *common.RedisConfig) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	// 测试连接
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %v", err)
	}

	globalClient = &Client{
		config: config,
		client: rdb,
	}
	return nil
}

// GetClient 获取全局Redis客户端
func GetClient() *Client {
	return globalClient
}

// Set 设置键值对
func (c *Client) Set(key string, value interface{}, expiration time.Duration) error {
	ctx := context.Background()
	return c.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取值
func (c *Client) Get(key string) (interface{}, error) {
	ctx := context.Background()
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("key not found")
		}
		return nil, err
	}
	return val, nil
}

// HSet 设置哈希字段
func (c *Client) HSet(key, field string, value interface{}) error {
	ctx := context.Background()
	return c.client.HSet(ctx, key, field, value).Err()
}

// HGet 获取哈希字段值
func (c *Client) HGet(key, field string) (interface{}, error) {
	ctx := context.Background()
	val, err := c.client.HGet(ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("key not found")
		}
		return nil, err
	}
	return val, nil
}

// Del 删除键
func (c *Client) Del(key string) error {
	ctx := context.Background()
	return c.client.Del(ctx, key).Err()
}

// Exists 检查键是否存在
func (c *Client) Exists(key string) bool {
	ctx := context.Background()
	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false
	}
	return count > 0
}

// SetWithExpiration 设置带过期时间的键值对
func (c *Client) SetWithExpiration(key string, value interface{}, seconds int) error {
	expiration := time.Duration(seconds) * time.Second
	return c.Set(key, value, expiration)
}

// 全局方法封装
func Set(key string, value interface{}, expiration time.Duration) error {
	if globalClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return globalClient.Set(key, value, expiration)
}

func Get(key string) (interface{}, error) {
	if globalClient == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}
	return globalClient.Get(key)
}

func HSet(key, field string, value interface{}) error {
	if globalClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return globalClient.HSet(key, field, value)
}

func HGet(key, field string) (interface{}, error) {
	if globalClient == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}
	return globalClient.HGet(key, field)
}

func Del(key string) error {
	if globalClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return globalClient.Del(key)
}

func Exists(key string) bool {
	if globalClient == nil {
		return false
	}
	return globalClient.Exists(key)
}

func SetWithExpiration(key string, value interface{}, seconds int) error {
	if globalClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return globalClient.SetWithExpiration(key, value, seconds)
}
