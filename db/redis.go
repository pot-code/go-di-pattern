package db

import (
	"log"

	"github.com/go-redis/redis"
)

type RedisDB struct {
	*redis.Client
}

func (rdb RedisDB) Constructor() *RedisDB {
	return &RedisDB{newRedisClient()}
}

func newRedisClient() *redis.Client {
	defer func() {
		if err := recover(); err != nil {
			log.Fatal(err)
		}
	}()
	client := redis.NewClient(&redis.Options{
		DB: 0,
	})
	_, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}
	return client
}
