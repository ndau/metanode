package state

import (
	"fmt"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

var redisAddr = "localhost:6379" // default redis server
var redisPass = "" 		 // no password
var redisDb   = 0  		 // default db

type SearchClient struct {
	client *redis.Client
}

// Search factory method.
func NewSearchClient() *SearchClient {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr, 
		Password: redisPass,
		DB:       redisDb,
	})

	searchClient := SearchClient{
		client: redisClient,
	}

	err := searchClient.testConnection()
	if err != nil {
		panic(err)
	}

	return &searchClient
}

// Helper function for creating consistent error messages from Search methods.
func errorMessage(method string, message string) string {
	return "SearchClient." + method + ": " + message
}

// Test connection to the index database service.
func (searchClient *SearchClient) testConnection() error {
	if searchClient == nil {
		return errors.New(errorMessage("testConnection", "searchClient cannot be nil"))
	}

	pong, err := searchClient.ping()
	if err != nil {
		msg := errorMessage(
			"testConnection",
			fmt.Sprintf("ping failed, is redis running at %s?", redisAddr))
		return errors.Wrap(err, msg)
	}

	if pong != "PONG" {
		return errors.New(errorMessage("testConnection", "Invalid pong"))
	}

	return nil
}

// Ping the index database service.
func (searchClient *SearchClient) ping() (string, error) {
	if searchClient == nil {
		err := errors.New(errorMessage("ping", "searchClient cannot be nil"))
		return "", err
	}

	if searchClient.client == nil {
		err := errors.New(errorMessage("ping", "searchClient.client cannot be nil"))
		return "", err
	}

	pong, err := searchClient.client.Ping().Result()

	return pong, err
}
