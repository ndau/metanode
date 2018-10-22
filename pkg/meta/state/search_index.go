package state

import (
	"fmt"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

var redisAddr = "localhost:6379" // default redis server
var redisPass = "" 		 // no password
var redisDb   = 0  		 // default db

type SearchIndex struct {
	client *redis.Client
}

func searchIndexPanic(method string, message string) {
	panic("SearchIndex." + method + ": " + message)
}

func NewSearchIndex() *SearchIndex {
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr, 
		Password: redisPass,
		DB:       redisDb,
	})

	searchIndex := SearchIndex{
		client: client,
	}

	return &searchIndex
}

func (searchIndex *SearchIndex) Connect() error {
	if searchIndex == nil {
		searchIndexPanic("Connect", "searchIndex cannot be nil")
	}

	// Not really "connecting" but this is a good way to test a redis connection.
	pong := searchIndex.ping()
	if pong == "PONG" {
		return nil
	}

	msg := fmt.Sprintf("Connection test failed, is redis running at %s?", redisAddr)
	return errors.New(msg)
}

func (searchIndex *SearchIndex) ping() string {
	if searchIndex == nil {
		searchIndexPanic("ping", "searchIndex cannot be nil")
	}

	if searchIndex.client == nil {
		searchIndexPanic("ping", "searchIndex.client cannot be nil")
	}

	pong, _ := searchIndex.client.Ping().Result()

	return pong
}
