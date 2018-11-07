package search

import (
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

var versionKey = "version" // Per-database key storing the database format version.
var heightKey = "height"   // Per-database key storing the height that we've indexed up to.

type SearchClient struct {
	client *redis.Client // Underlying redis database client.
	height uint64        // The blockchain height that we've indexed up to.
}

// Factory method.
// The address should contain ip and port with no "http://", e.g. "localhost:6379".
// Pass in a version number for your client.  Start it at zero.  If you later increment it,
// the client will wipe the database and require reindexing.
func NewSearchClient(address string, version int) (search *SearchClient, err error) {
	if version < 0 {
		err = errors.New("SearchClient version must be non-negative")
		return nil, err
	}

	client := redis.NewClient(&redis.Options{
		Addr: address,
	})

	search = &SearchClient{
		client: client,
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
	return "SearchClient." + method + ": " + message
}

// Test connection to the redis service.
func (search *SearchClient) testConnection(address string) error {
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
func (search *SearchClient) processSearchVersion(version int) (err error) {
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

		// Edge case: leave the default of search.height = 0 if nothing was stored.
		if len(heightString) != 0 {
			search.height, err = strconv.ParseUint(heightString, 10, 64)
			if err != nil {
				return err
			}
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
func (search *SearchClient) testValidity(method string) error {
	if search == nil {
		err := errors.New(errorMessage(method, "search cannot be nil"))
		return err
	}

	if search.client == nil {
		err := errors.New(errorMessage(method, "search.client cannot be nil"))
		return err
	}

	return nil
}

// Call this any time you index something at a given blockchain height.
// It's also acceptable to call this once after an initial scan.
// It will make the next scan-on-launch only index blocks down to this height.
func (search *SearchClient) SetHeight(height uint64) (err error) {
	err = search.testValidity("SetHeight")
	if err != nil {
		return err
	}

	if height > search.height {
		search.height = height
		err = search.Set(heightKey, height)
		if err != nil {
			return err
		}
	}

	return nil
}

// Get the high water mark (height) we've indexed to so far.
func (search *SearchClient) GetHeight() uint64 {
	return search.height
}

// Wrapper for redis PING.
func (search *SearchClient) Ping() error {
	err := search.testValidity("Ping")
	if err != nil {
		return err
	}

	result, err := search.client.Ping().Result()
	if err != nil {
		return err
	}
	if result != "PONG" {
		return errors.New(errorMessage("Ping", fmt.Sprintf("expected 'PONG', got '%s'", result)))
	}

	return nil
}

// Wrapper for redis FLUSHDB.
func (search *SearchClient) FlushDB() error {
	err := search.testValidity("FlushDB")
	if err != nil {
		return err
	}

	result, err := search.client.FlushDB().Result()
	if err != nil {
		return err
	}
	if result != "OK" {
		return errors.New(errorMessage("FlushDB", fmt.Sprintf("expected 'OK', got '%s'", result)))
	}

	return nil
}

// Wrapper for redis SET with no expiration.
func (search *SearchClient) Set(key string, value interface{}) error {
	err := search.testValidity("Set")
	if err != nil {
		return err
	}

	result, err := search.client.Set(key, value, 0).Result()
	if err != nil {
		return err
	}
	if result != "OK" {
		return errors.New(errorMessage("Set", fmt.Sprintf("expected 'OK', got '%s'", result)))
	}

	return nil
}

// Wrapper for redis GET.  Returns empty string (not nil) if the key doesn't exist.
func (search *SearchClient) Get(key string) (string, error) {
	err := search.testValidity("Get")
	if err != nil {
		return "", err
	}

	result, err := search.client.Get(key).Result()
	if err == redis.Nil {
		return "", nil
	}

	return result, err
}

// Wrapper for redis HSET.  Returns true for new fields, false if the field already exists.
func (search *SearchClient) HSet(key, field string, value interface{}) (bool, error) {
	err := search.testValidity("HSet")
	if err != nil {
		return false, err
	}

	return search.client.HSet(key, field, value).Result()
}

// Wrapper for redis HGET.
func (search *SearchClient) HGet(key, field string) (string, error) {
	err := search.testValidity("HGet")
	if err != nil {
		return "", err
	}

	return search.client.HGet(key, field).Result()
}

// Wrapper for redis HGETALL.
func (search *SearchClient) HGetAll(key string) (map[string]string, error) {
	err := search.testValidity("HGetAll")
	if err != nil {
		return nil, err
	}

	return search.client.HGetAll(key).Result()
}

// Wrapper for redis ZADD.  Returns the number of elements added.
func (search *SearchClient) ZAdd(
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
	return search.client.ZAdd(key, member).Result()
}

// Wrapper for redis full-iteration ZSCAN with wildcard match.
func (search *SearchClient) ZScan(
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

		// Redis documents that 10 is the default count for all scan commands.
		results, cursor, err = search.client.ZScan(key, cursor, "", 10).Result()
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
