package search

import (
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

// Redis config.
var redisAddr = "localhost:6379" // default redis server
var redisPass = "" 		 // no password

// Constants.
var databasesKey = "databases" // Search system variable that stores the redis db numbers we use.
var versionKey = "version" // Per-database key that stores the database format version.
var heightKey  = "height"  // Per-database key that stores the height that we've indexed up to.

type SearchClient struct {
	client *redis.Client // Underlying redis database client.
	Height uint64        // The blockchain height that we've indexed up to.
}

// Search factory method.
// Applications must supply a unique name so that we can map it to a unique redis database number.
// Pass in a version number for your client.  Start it at zero.  If you later increment it,
// the client will wipe the current database associated with the given name.
func NewSearchClient(name string, version int) *SearchClient {
	if name == "" {
		panic("SearchClient name must be non-empty")
	}

	if version < 0 {
		panic("SearchClient version must be non-negative")
	}

	systemClient := redis.NewClient(&redis.Options{
		Addr:      redisAddr, 
		Password:  redisPass,
		DB:        0, // We start with the zero'th db to grab search system variables.
	})

	search := &SearchClient{
		client: systemClient,
		Height: 0,
	}

	// Test connection to the search system client.
	err := search.testConnection()
	if err != nil {
		// FIXME: Figure out how to run in a dormant mode for tests, but panic otherwise.
		//panic(err)
		return nil
	}

	// With the system client active, get the redis database index.
	var dbIndex int64
	dbIndex, err = search.getDatabaseIndex(name)
	if err != nil {
		panic(err)
	}

	// No longer need the system client.
	systemClient.Close()

	// Now that we know the database index we can create and select the main client.
	// TODO: Would be nice to use `SELECT dbIndex` instead, not close/create another client.
	search.client = redis.NewClient(&redis.Options{
		Addr:      redisAddr, 
		Password:  redisPass,
		DB:        int(dbIndex),
	})

	// Test connection to the main client.
	err = search.testConnection()
	if err != nil {
		panic(err)
	}

	// Flush the database if the caller has incremented the version since the last time.
	err = search.processSearchVersion(version)
	if err != nil {
		panic(err)
	}

	return search
}

// Helper function for creating consistent error messages from Search methods.
func errorMessage(method string, message string) string {
	return "SearchClient." + method + ": " + message
}

// Test connection to the redis service.
func (search *SearchClient) testConnection() error {
	pong, err := search.Ping()
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

// Helper function for getting the redis database number from the system search client.
// Will assign and return an available database number if there isn't one associated with name.
// The system search client must be selected (set into search.client) before calling.
func (search *SearchClient) getDatabaseIndex(name string) (dbIndex int64, err error) {
	// Grab the database numbers map from the search system database.
	var databases map[string]string
	databases, err = search.HGetAll(databasesKey)
	if err != nil {
		return -1, err
	}

	// See if we have the given name in the map already.
	if dbString, ok := databases[name]; ok {
		// Use the database number from the map.
		dbIndex, err = strconv.ParseInt(dbString, 10, 32)
		if err != nil {
			return -1, err
		}
	} else {
		// Find the next available database number for us to use.
		dbIndex = 0
		for _, indexString := range databases {
			var index int64
			index, err = strconv.ParseInt(indexString, 10, 32)
			if err != nil {
				return -1, err
			}
			if index > dbIndex {
				dbIndex = index
			}
		}
		dbIndex++

		// Save it in the map for next time.
		_, err = search.HSet(databasesKey, name, dbIndex)
		if err != nil {
			return -1, err
		}
	}

	return dbIndex, nil
}

// Helper function for flushing the database if the version number has been incremented since the
// last time we used this client.  Also grabs the height we've indexed up to.
// The main search client must be selected (set into search.client) before calling.
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
	if versionString != "" {
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

		// Edge case: leave the default of search.Height = 0 if nothing was stored.
		if heightString != "" {
			search.Height, err = strconv.ParseUint(heightString, 10, 64)
			if err != nil {
				return err
			}
		}
	} else {
		// The version was incremented from what we have stored in the database.
		// Wipe the database at the currently selected db index.
		_, err = search.FlushDB()
		if err != nil {
			return err
		}

		// Set the height for completeness.
		// We leave the default of search.Height = 0 in this case.
		_, err = search.Set(heightKey, 0)
		if err != nil {
			return err
		}

		// Store the new version.
		_, err = search.Set(versionKey, version)
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

	if height > search.Height {
		search.Height = height
		_, err = search.Set(heightKey, height)
		if err != nil {
			return err
		}
	}

	return nil
}
	
// Wrapper for redis PING.  Returns "PONG" on success.
func (search *SearchClient) Ping() (result string, err error) {
	err = search.testValidity("Ping")
	if err != nil {
		return "", err
	}

	result, err = search.client.Ping().Result()

	return result, err
}

// Wrapper for redis FLUSHDB.  Returns "OK" on success.
func (search *SearchClient) FlushDB() (result string, err error) {
	err = search.testValidity("FlushDB")
	if err != nil {
		return "", err
	}

	result, err = search.client.FlushDB().Result()

	return result, err
}

// Wrapper for redis SET with no expiration.  Returns "OK" on success.
func (search *SearchClient) Set(key string, value interface{}) (result string, err error) {
	err = search.testValidity("Set")
	if err != nil {
		return "", err
	}

	result, err = search.client.Set(key, value, 0).Result()

	return result, err
}

// Wrapper for redis GET.  Returns empty string (not nil) if the key doesn't exist.
func (search *SearchClient) Get(key string) (result string, err error) {
	err = search.testValidity("Get")
	if err != nil {
		return "", err
	}

	result, err = search.client.Get(key).Result()
	if err == redis.Nil {
		return "", nil
	}

	return result, err
}

// Wrapper for redis HSET.  Returns true for new fields, false if the field already exists.
func (search *SearchClient) HSet(key, field string, value interface{}) (result bool, err error) {
	err = search.testValidity("HSet")
	if err != nil {
		return false, err
	}

	result, err = search.client.HSet(key, field, value).Result()

	return result, err
}

// Wrapper for redis HGET.
func (search *SearchClient) HGet(key, field string) (result string, err error) {
	err = search.testValidity("HGet")
	if err != nil {
		return "", err
	}

	result, err = search.client.HGet(key, field).Result()

	return result, err
}

// Wrapper for redis HGETALL.
func (search *SearchClient) HGetAll(key string) (result map[string]string, err error) {
	err = search.testValidity("HGetAll")
	if err != nil {
		return nil, err
	}

	result, err = search.client.HGetAll(key).Result()

	return result, err
}

// Wrapper for redis ZADD.  Returns the number of elements added.
func (search *SearchClient) ZAdd(
	key string, score float64, value string,
) (count int64, err error) {
	err = search.testValidity("ZAdd")
	if err != nil {
		return 0, err
	}

	member := redis.Z {
		Score: score,
		Member: value,
	}

	count, err = search.client.ZAdd(key, member).Result()
	if err != nil {
		return count, err
	}

	return count, nil
}

// Wrapper for redis full-iteration ZSCAN with wildcard match.
func (search *SearchClient) ZScan(
	key string,
	cb func(value string, score float64) error,
) (err error) {
	err = search.testValidity("ZScan")
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
