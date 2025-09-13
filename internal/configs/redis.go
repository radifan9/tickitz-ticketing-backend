package configs

import (
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

func InitRDB() (*redis.Client, error) {
	redisUser := os.Getenv("REDIS_USER")
	redisPass := os.Getenv("REDIS_PASSWORD")
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Username: redisUser,
		Password: redisPass,
		DB:       0,
	})

	return rdb, nil
}
