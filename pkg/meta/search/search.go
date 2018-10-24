package search

import (
	"fmt"
	"strconv"

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

	pong, err := searchClient.Ping()
	if err != nil {
		msg := errorMessage(
			"testConnection",
			fmt.Sprintf("Ping failed, is redis running at %s?", redisAddr))
		return errors.Wrap(err, msg)
	}

	if pong != "PONG" {
		return errors.New(errorMessage("testConnection", "Invalid pong"))
	}

	return nil
}

// Test whether the search client is ready to have redis commands run on it.
func (searchClient *SearchClient) testValidity(method string) error {
	if searchClient == nil {
		err := errors.New(errorMessage(method, "searchClient cannot be nil"))
		return err
	}

	if searchClient.client == nil {
		err := errors.New(errorMessage(method, "searchClient.client cannot be nil"))
		return err
	}

	return nil
}

// Wrapper for redis PING.  Returns "PONG" on success.
func (searchClient *SearchClient) Ping() (result string, err error) {
	err = searchClient.testValidity("Ping")
	if err != nil {
		return "", err
	}

	result, err = searchClient.client.Ping().Result()

	return result, err
}

// Wrapper for redis FLUSHDB.  Returns "OK" on success.
func (searchClient *SearchClient) FlushDB() (result string, err error) {
	err = searchClient.testValidity("FlushDB")
	if err != nil {
		return "", err
	}

	result, err = searchClient.client.FlushDB().Result()

	return result, err
}

// Wrapper for redis ZADD.
func (searchClient *SearchClient) ZAdd(key string, score float64, value string) (err error) {
	err = searchClient.testValidity("ZAdd")
	if err != nil {
		return err
	}

	member := redis.Z {
		Score: score,
		Member: value,
	}

	var count int64
	count, err = searchClient.client.ZAdd(key, member).Result()
	if err != nil {
		return err
	}

	// If we didn't add exactly one element, we consider it a failure.
	if count != 1 {
		err = errors.New(errorMessage(
			"ZAdd",
			fmt.Sprintf("Unable to ZADD %s=%s with score=%f", key, value, score),
		))
		return err
	}

	return nil
}

// Wrapper for redis full-iteration ZSCAN with wildcard match.
func (searchClient *SearchClient) ZScan(
	key string,
	cb func(value string, score float64) error,
) (err error) {
	err = searchClient.testValidity("ZScan")
	if err != nil {
		return err
	}

	cursor := uint64(0)
	for {
		var results []string

		// Redis documents that 10 is the default count for all scan commands.
		results, cursor, err = searchClient.client.ZScan(key, cursor, "", 10).Result()
		if err != nil {
			return err
		}

		len := len(results)
		for i := 0; i < len; i += 2 {
			value := results[i]
			var score float64
			score, err = strconv.ParseFloat(results[i + 1], 64)
			if err != nil {
				return err
			}
			err = cb(value, score)
			if err != nil {
				return err
			}
		}

		if cursor == 0 {
			break
		}
	}

	return nil
}
