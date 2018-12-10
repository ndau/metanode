package search

// Blockchain-independent implementation for date range indexing and searching.

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Date range interval is how many seconds between snapshot we take of the blockchain height.
// This should be an integer that divides the number of seconds in a day evenly.  i.e. it must be
// a divisor of 86400 = 24 * 60 * 60 = 2^7 * 3^3 * 5^2.
//
// Using 3600 for this value means "take a height snapshot every hour, starting at midnight UTC".
// Using 8640 for this value means "take 10 snapshots per day", or "once every 8,640 seconds".
// Using 86400 for this value means "take one snapshot every day at midnight".
// In any case, we take a snapshot at midnight UTC.
//
// "Daily" (dateRangeInterval = 86400) is the minimum snapshot frequency.
// "Every second" (dateRangeInterval = 1) is the max (but very wasteful and will bloat the index).
//
// Since we use UTC, we avoid common daylight savings time hassles and can assume that "every day
// is made up of exactly 86400 seconds".  As for leap seconds, the assumption is that when a leap
// second occurs, time stands still for that second.  There is no indication when referring to
// timestamps in UTC that a leap second has occurred, so we do not have to account for it.  This
// means that we can safely assume that a single day's worth of seconds is an exact multiple of
// dateRangeInterval.  This is important, as some of our timestamp arithmetic adds or subtracts
// this many seconds from one snapshot timestamp to compute an adjacent snapshot timestamp.
// It is true that when a leap second occurs, the day has 86401 seconds in it.  But since we use
// UTC timestamps, and convert to/from seconds-past-midnight, we never notice the leap second.
const dateRangeInterval = 86400

// As we're taking snapshots, we index this for the key of the next snapshot time.  This is so we
// have something indexed there to signal that it's a valid snapshot time key, but we haven't yet
// taken the snapshot.  Once we do, we'll replace the flag with the actual blockchain height.
const nextHeightFlag = "*"

// As a way of grouping keys, we use this prefix for date range height snapshot key names.
// We also use this to store the last snapshot key we indexed.  Useful for blockchains that get
// update infrequently.  Primarily to avoid an infinite loop when filling in missing snapshots.
const dateRangeToHeightSearchKeyPrefix = "d:"

// Number of nanoseconds in a second.
const oneSecond = 1e9

// Return the number of seconds past midnight of the given time, ignoring nanoseconds.
func secondsAfterMidnight(t time.Time) int {
	return t.Hour() * 3600 + t.Minute() * 60 + t.Second()
}

// Return the given time, truncated to midnight, then with the given seconds added to it.
func timeAfterMidnight(t time.Time, seconds int) time.Time {
	hours := seconds / 3600
	seconds -= hours * 3600

	minutes := seconds / 60
	seconds -= minutes * 60

	return time.Date(t.Year(), t.Month(), t.Day(), hours, minutes, seconds, 0, t.Location())
}

// Return the floor of the given time in seconds to the nearest day interval constant.
func floorSeconds(seconds int) int {
	return (seconds / dateRangeInterval) * dateRangeInterval
}

// Return the ceiling of the given time in seconds to the nearest day interval constant.
func ceilSeconds(seconds int) int {
	return floorSeconds(seconds + dateRangeInterval - 1)
}

// Truncate to appropriate snapshot interval using the given trunc method.
func truncTime(t time.Time, truncMethod func(int) int) time.Time {
	seconds := secondsAfterMidnight(t)
	seconds = truncMethod(seconds)
	return timeAfterMidnight(t, seconds)
}

// Format the given time into a date key that we index and search on for date range queries.
// The time should already be trunctated to a multiple of dateRangeInterval seconds past midnight.
func formatDateRangeToHeightSearchKey(truncatedTime time.Time) string {
	key := truncatedTime.Format(time.RFC3339)

	// In practice, this was never needed.  But for sanity, let's make sure the keys has the
	// expected length of 20.  That's "YYYY-MM-DDTHH:MM:SSZ" with no nanoseconds on it.
	// We don't want to index any nanoseconds.
	if len(key) > 20 {
		// Strip off all the nanoseconds (and the Z) then put the Z back on.
		key = key[:19] + "Z"
	}

	return dateRangeToHeightSearchKeyPrefix + key
}

// Return the blockchain height from the index for the given timestamp.
// truncMethod is for rounding the time to a multiple of dateRangeInterval seconds past midnight.
func (search *Client) getHeightFromTime(
	timeParam string, truncMethod func(int) int,
) (uint64, error) {
	t, err := time.Parse(time.RFC3339, timeParam)
	if err != nil {
		return 0, err
	}

	key := formatDateRangeToHeightSearchKey(truncTime(t, truncMethod))
	value, err := search.Get(key)
	if err != nil {
		return 0, err
	}

	if value == "" {
		// This won't happen if the caller makes sure to pass in a time that is within the
		// range of possible blockchain block timestamps.
		return 0, fmt.Errorf("Could not find %s in the index for %s", key, timeParam)
	}

	// We always keep the next snapshot key indexed with a special flag, so that it doesn't
	// come back empty when we ask for it.  We know it's a valid key we asked for, but haven't
	// taken the height snapshot for it yet.
	if value == nextHeightFlag {
		return search.height, nil
	}

	return strconv.ParseUint(value, 10, 64)
}

// SearchDateRange returns the first and last block heights for the given ISO-3339 date range.
// The first is inclusive, the last is exclusive.
func (search *Client) SearchDateRange(first, last string) (uint64, uint64, error) {
	// Floor the first time to the nearest day interval constant.
	firstHeight, err := search.getHeightFromTime(first, floorSeconds)
	if err != nil {
		return 0, 0, err
	}

	var lastHeight uint64
	if last == "" {
		// An empty last time parameter means the query is to include all trailing blocks.
		// The search.height stores the next block height to index, so this creates an exclusive
		// last height, which is what we want.
		lastHeight = search.height
	} else {
		// We floor the first time to the nearest day interval, but ceil the last time.
		// That way, you get at least what was asked for (vs flooring both, for example).
		lastHeight, err = search.getHeightFromTime(last, ceilSeconds)
		if err != nil {
			return 0, 0, err
		}
	}	

	return firstHeight, lastHeight, nil
}

// Helper function for getting the two snapshot times surrounding the given time.
func getPrevAndNextTimes(t time.Time) (prevTime, nextTime time.Time) {
	prevTime = truncTime(t, floorSeconds)
	nextTime = truncTime(prevTime.Add(time.Duration(oneSecond)), ceilSeconds)
	return
}

// Helper function for initilizeing the date range index at height zero.
// Called by base search client implementation when it detects a fresh/blank index.
func (search *Client) initializeDateRangeIndex() (err error) {
	// Current use case is to be called when the index is empty.
	t := time.Now()
	height := 0

	prevTime, nextTime := getPrevAndNextTimes(t)

	prevKey := formatDateRangeToHeightSearchKey(prevTime)
	err = search.Set(prevKey, height)
	if err != nil {
		return err
	}

	nextKey := formatDateRangeToHeightSearchKey(nextTime)
	err = search.Set(nextKey, nextHeightFlag)
	if err != nil {
		return err
	}

	// Save off the next-height snapshot key in the special prefix-only key.
	err = search.Set(dateRangeToHeightSearchKeyPrefix, nextKey)
	if err != nil {
		return err
	}

	return nil
}

// IndexDateToHeight will index all necessary date-to-height keys back in time to the latest one
// we've indexed, using the given date and height.  Typically this function will only need to do
// work once every dateRangeInterval seconds.  But if there are long periods of block inactivity,
// this function will fill in all missing date-to-height keys up to the given block time.
// The given block height must be > 0, which is guaranteed if it comes from Tendermint.
func (search *Client) IndexDateToHeight(
	blockTime time.Time, blockHeight uint64,
) (updateCount int, insertCount int, err error) {
	updateCount = 0
	insertCount = 0

	// Ignore invalid block times.
	if blockTime.Year() <= 1 {
		return updateCount, insertCount, nil
	}

	// We already have the initial 0-height date snapshots indexed.  This is an edge case.
	// We need to subtract one from the given block height below, so exit early if we can't.
	// We'll continue normally when we're called again for block height 1.
	if blockHeight == 0 {
		return updateCount, insertCount, nil
	}

	prevTime, nextTime := getPrevAndNextTimes(blockTime)

	nextKey := formatDateRangeToHeightSearchKey(nextTime)
	nextValue, err := search.Get(nextKey)
	if err != nil {
		return updateCount, insertCount, err
	}

	// Common case: all snapshots up-to-date.
	if nextValue == nextHeightFlag {
		return updateCount, insertCount, nil
	}

	// There shouldn't be anything there.  If there is, something went wrong and it's an error.
	// Let's not fail in this case.  Let's treat it as if there's a valid block height there.
	// This might happen if somehow a new block's timestamp is earlier than a previou's block's
	// timestamp that we've already indexed.  Since we use UTC timestamps, this should never
	// happen.  There is no daylight savings time, for example, to make the clock go backwards.
	// But maybe the system we're running on has had its clock altered.  That'll screw things up
	// but we don't want to let it cause errors for us.  We'll just have to wait until the clock
	// catches up to previously indexed timestamps, and then we'll continue normally from there.
	if nextValue != "" {
		return updateCount, insertCount, nil
	}

	// Need to fill in new snapshots.  Start with inserting the new nextHeightFlag snapshot.
	err = search.Set(nextKey, nextHeightFlag)
	if err != nil {
		return updateCount, insertCount, err
	}
	insertCount++

	// These next steps ensure that initializeDateRangeIndex() has been called and that the
	// genesis snapshot hasn't been lost, preventing the chance of an infinite loop below.
	var specialTime time.Time
	specialKey, err := search.Get(dateRangeToHeightSearchKeyPrefix)
	if err != nil {
		return updateCount, insertCount, err
	}
	if strings.Index(specialKey, dateRangeToHeightSearchKeyPrefix) != 0 {
		// We could fail here, but since block timestamps are currently not something that can be
		// initially indexed by some node apps, we haven't incremented the index version when the
		// date-to-height index was implemented.  So we silently start up the index on the next
		// blocked that is incrementally indexed.  It just means searches for earlier dates will
		// return zero results.  We have a ticket to write an external initial indexing app for
		// each blockchain, and once that exists, we'll bump the index version and this case will
		// no longer need to be handled.  It'll become a harmless sanity check at that point.
		specialTime = prevTime
	} else {
		specialTimestamp := specialKey[len(dateRangeToHeightSearchKeyPrefix):]
		specialTime, err = time.Parse(time.RFC3339, specialTimestamp)
		if err != nil {
			return updateCount, insertCount, err
		}
		specialValue, err := search.Get(specialKey)
		if err != nil {
			return updateCount, insertCount, err
		}
		if specialValue != nextHeightFlag {
			return updateCount, insertCount, fmt.Errorf(
				"Unexpected special value '%s' in key %s", specialValue, specialKey)
		}
	}
	err = search.Set(dateRangeToHeightSearchKeyPrefix, nextKey)
	if err != nil {
		return updateCount, insertCount, err
	}
	updateCount++

	// Save this off in case we have to fill in missing snapshots.
	lastBlockHeight := blockHeight - 1

	// Common case: the new block time is almost certainly not right on a snapshot boundary...
	if !blockTime.Equal(prevTime) {
		// ...so the height at the snapshot point is one less than the new block's height.
		blockHeight = lastBlockHeight
	}

	// Initial conditions for the "effective for-loop" below.
	prevKey := formatDateRangeToHeightSearchKey(prevTime)
	prevValue, err := search.Get(prevKey)
	if err != nil {
		return updateCount, insertCount, err
	}

	// Fill in missing snapshots.
	// Common case: prevValue will be nextHeightFlag and we'll iterate once.
	for prevValue == "" || prevValue == nextHeightFlag {
		err = search.Set(prevKey, blockHeight)
		if err != nil {
			return updateCount, insertCount, err
		}

		if prevValue == "" {
			insertCount++
		} else {
			// We replaced the nextHeightFlag, so it's an update, not an insert.
			updateCount++

			// The next iteration would break since we'd find a non-empty, valid height.
			// But it's more optimal to exit early here.
			break
		}

		if !prevTime.After(specialTime) {
			// Prevent infinite loop in case something went wrong in the index.
			break
		}

		prevTime = prevTime.Add(time.Duration(-dateRangeInterval * oneSecond))
		prevKey = formatDateRangeToHeightSearchKey(prevTime)
		prevValue, err = search.Get(prevKey)
		if err != nil {
			return updateCount, insertCount, err
		}

		// Since we found a hole, we want to index the last block height, not the current.
		// We likely already set this, but this handles the edge case if we didn't.
		blockHeight = lastBlockHeight
	}

	return updateCount, insertCount, nil
}
