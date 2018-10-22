package state

import (
	"github.com/go-redis/redis"
)

type SearchIndex struct {
	client *redis.Client
}

func RedisExample() *SearchIndex {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // default redis server
		Password: "",               // no password
		DB:       0,                // default db
	})

	searchIndex := SearchIndex{
		client: client,
	}

	return &searchIndex
}
