package search

import (
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

const versionKey = "version" // Per-database key storing the database format version.
const heightKey = "height"   // Per-database key storing the height that we've indexed up to.

// Redis documents that 10 is the default count for all scan commands.
const scanCount = int64(10)

// Client manages a redis.Client for use with indexing and searching within a node.
type Client struct {
	redis  *redis.Client // Underlying redis database client.
	height uint64        // The blockchain height that we've indexed up to, but not including.
}

// NewClient is a factory method for Client.
// The address should contain ip and port with no "http://", e.g. "localhost:6379".
// Pass in a version number for your client.  Start it at zero.  If you later increment it,
// the client will wipe the database and require reindexing.
func NewClient(address string, version int) (search *Client, err error) {
	if version < 0 {
		err = errors.New("Client version must be non-negative")
		return nil, err
	}

	redis := redis.NewClient(&redis.Options{
		Addr: address,
	})

	search = &Client{
		redis:  redis,
		height: 0,
	}

	err = search.testConnection(address)
	if err != nil {
		return nil, err
	}

	err = search.processSearchVersion(version)
	if err != nil {
		return nil, err
	}

	return search, nil
}

// Helper function for creating consistent error messages from Search methods.
func errorMessage(method string, message string) string {
	return "Client." + method + ": " + message
}

// Test connection to the redis service.
func (search *Client) testConnection(address string) error {
	err := search.Ping()
	if err != nil {
		msg := errorMessage(
			"testConnection",
			fmt.Sprintf("Ping failed, is redis running at %s?", address))
		return errors.Wrap(err, msg)
	}

	return nil
}

// Helper function for flushing the database if the version number has been incremented since
// the last time we used this client.  Also grabs the height we had indexed up to before.
func (search *Client) processSearchVersion(version int) (err error) {
	// Use -1 by default to trigger setting a new version as a search system variable.
	existingVersion := int64(-1)

	// Get the existing search version number.
	var versionString string
	versionString, err = search.Get(versionKey)
	if err != nil {
		return err
	}

	// Edge case: leave the default existingVersion = -1 if nothing was stored.
	if len(versionString) != 0 {
		existingVersion, err = strconv.ParseInt(versionString, 10, 32)
		if err != nil {
			return err
		}
	}

	if existingVersion >= int64(version) {
		// We support this version, get the height we've indexed up to.
		var heightString string
		heightString, err = search.Get(heightKey)
		if err != nil {
			return err
		}

		// Leave the default of search.height = 0 if nothing was stored.
		if len(heightString) != 0 {
			height, err := strconv.ParseUint(heightString, 10, 64)
			if err != nil {
				return err
			}
			search.height = height
		}
	} else {
		// The version was incremented from what we have stored in the database.
		// Wipe the database.
		err = search.FlushDB()
		if err != nil {
			return err
		}

		// Set the height for completeness.
		// We leave the default of search.height = 0 in this case.
		err = search.Set(heightKey, 0)
		if err != nil {
			return err
		}

		// Store the new version.
		err = search.Set(versionKey, version)
		if err != nil {
			return err
		}
	}

	return nil
}

// Test whether the search client is ready to have redis commands run on it.
func (search *Client) testValidity(method string) error {
	if search == nil {
		err := errors.New(errorMessage(method, "search cannot be nil"))
		return err
	}

	if search.redis == nil {
		err := errors.New(errorMessage(method, "search.redis cannot be nil"))
		return err
	}

	return nil
}

// SetNextHeight saves the given height in the database as a high water mark.
// Call this after you've indexed something at a given blockchain height.
// It's also acceptable to call this once after an initial scan.
// It will make the next scan-on-launch index blocks down to, and including, this height.
func (search *Client) SetNextHeight(height uint64) (err error) {
	err = search.testValidity("SetHeight")
	if err != nil {
		return err
	}

	if height > search.height {
		err = search.Set(heightKey, height)
		if err != nil {
			return err
		}
		search.height = height
	}

	return nil
}

// GetNextHeight gets the high water mark (height) we've indexed up to, but not including it.
func (search *Client) GetNextHeight() uint64 {
	return search.height
}

// Ping is a wrapper for redis PING.
func (search *Client) Ping() error {
	err := search.testValidity("Ping")
	if err != nil {
		return err
	}

	result, err := search.redis.Ping().Result()
	if err != nil {
		return err
	}
	if result != "PONG" {
		return errors.New(errorMessage("Ping", fmt.Sprintf("expected 'PONG', got '%s'", result)))
	}

	return nil
}

// FlushDB is a wrapper for redis FLUSHDB.
func (search *Client) FlushDB() error {
	err := search.testValidity("FlushDB")
	if err != nil {
		return err
	}

	result, err := search.redis.FlushDB().Result()
	if err != nil {
		return err
	}
	if result != "OK" {
		return errors.New(errorMessage("FlushDB", fmt.Sprintf("expected 'OK', got '%s'", result)))
	}

	return nil
}

// Set is a wrapper for redis SET with no expiration.
func (search *Client) Set(key string, value interface{}) error {
	err := search.testValidity("Set")
	if err != nil {
		return err
	}

	result, err := search.redis.Set(key, value, 0).Result()
	if err != nil {
		return err
	}
	if result != "OK" {
		return errors.New(errorMessage("Set", fmt.Sprintf("expected 'OK', got '%s'", result)))
	}

	return nil
}

// Get is a wrapper for redis GET.  Returns empty string (not nil) if the key doesn't exist.
func (search *Client) Get(key string) (string, error) {
	err := search.testValidity("Get")
	if err != nil {
		return "", err
	}

	result, err := search.redis.Get(key).Result()
	if err == redis.Nil {
		return "", nil
	}

	return result, err
}

// HSet is a wrapper for redis HSET.  Returns true for new fields, false if field already exists.
func (search *Client) HSet(key, field string, value interface{}) (bool, error) {
	err := search.testValidity("HSet")
	if err != nil {
		return false, err
	}

	return search.redis.HSet(key, field, value).Result()
}

// HGet is a wrapper for redis HGET.
func (search *Client) HGet(key, field string) (string, error) {
	err := search.testValidity("HGet")
	if err != nil {
		return "", err
	}

	return search.redis.HGet(key, field).Result()
}

// HGetAll is a wrapper for redis HGETALL.
func (search *Client) HGetAll(key string) (map[string]string, error) {
	err := search.testValidity("HGetAll")
	if err != nil {
		return nil, err
	}

	return search.redis.HGetAll(key).Result()
}

// SAdd is a wrapper for redis SADD.  Returns the number of elements added.
func (search *Client) SAdd(
	key string, value string,
) (int64, error) {
	err := search.testValidity("SAdd")
	if err != nil {
		return 0, err
	}

	return search.redis.SAdd(key, value).Result()
}

// SScan is a wrapper for redis full-iteration SSCAN with wildcard match.
func (search *Client) SScan(
	key string,
	cb func(value string) error,
) error {
	err := search.testValidity("SScan")
	if err != nil {
		return err
	}

	cursor := uint64(0)
	for {
		var results []string

		results, cursor, err = search.redis.SScan(key, cursor, "", scanCount).Result()
		if err != nil {
			return err
		}

		for _, value := range results {
			err = cb(value)
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

// ZAdd is a wrapper for redis ZADD.  Returns the number of elements added.
func (search *Client) ZAdd(
	key string, score float64, value string,
) (int64, error) {
	err := search.testValidity("ZAdd")
	if err != nil {
		return 0, err
	}

	member := redis.Z{
		Score:  score,
		Member: value,
	}
	return search.redis.ZAdd(key, member).Result()
}

// ZScan is a wrapper for redis full-iteration ZSCAN with wildcard match.
func (search *Client) ZScan(
	key string,
	cb func(value string, score float64) error,
) error {
	err := search.testValidity("ZScan")
	if err != nil {
		return err
	}

	cursor := uint64(0)
	for {
		var results []string

		results, cursor, err = search.redis.ZScan(key, cursor, "", scanCount).Result()
		if err != nil {
			return err
		}

		// ZSCAN returns values and scores on their own rows, iterate two at a time.
		len := len(results)
		for i := 0; i < len; i += 2 {
			value := results[i]
			var score float64
			score, err = strconv.ParseFloat(results[i+1], 64)
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
