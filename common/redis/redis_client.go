package redis

import (
	"sync"

	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

var (
	redisClient *redis.Client
	initOnce    sync.Once
)

// GetClient singleton redis.Client
func GetClient() *redis.Client {
	initOnce.Do(func() {
		//连接服务器
		redisClient = redis.NewClient(&redis.Options{
			Addr:     viper.GetString("redis.addr"),     // use default Addr
			Password: viper.GetString("redis.password"), // no password set
			DB:       viper.GetInt("redis.db"),          // use default DB
		})
	})
	if redisClient == nil {
		panic("redisdb is nil")
	}

	return redisClient
}
