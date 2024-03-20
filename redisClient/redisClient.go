package redisClient

import (
	"github.com/redis/go-redis/v9"
)

func Init() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "docker", // no password set
		DB:       0,        // use default DB
	})

	return client
}
