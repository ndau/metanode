package search

// Blockchain-independent implementation for date range indexing and searching.

import (
	"fmt"
	"strconv"
	"time"
)

// Date range interval is how many seconds between snapshot we take of the blockchain height.
// This should be an integer that divides the number of seconds in a day evenly.  i.e. it must be
// a divisor of 86400 = 24 * 60 * 60 = 2^7 * 3^3 * 5^2.
// Using 3600 for this value means "take a height snapshot every hour, starting at midnight UTC".
// Using 8640 for this value means "take 10 snapshots per day", or "once every 8,640 seconds".
// Using 86400 for this value means "take one snapshot every day at midnight".
// In any case, we take a snapshot at midnight UTC.
// "Daily" (dateRangeInterval = 86400) is the minimum snapshot frequency.
// "Every second" (dateRangeInterval = 1) is the max (but very wasteful and will bloat the index).
const dateRangeInterval = 86400

// As we're taking snapshots, we index this for the key of the next snapshot time.  This is so we
// have something indexed there to signal that it's a valid snapshot time key, but we haven't yet
// taken the snapshot.  Once we do, we'll replace the flag with the actual blockchain height.
const nextHeightFlag = "*"

// As a way of grouping keys, we use this prefix for date range height snapshot key names.
const dateRangeToHeightSearchKeyPrefix = "d:"

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

// Format the given time into a date key that we index and search on for date range queries.
func formatDateRangeToHeightSearchKey(t time.Time) string {
	key := t.Format(time.RFC3339)

	// In practice, this was never needed.  But for sanity, let's make sure the keys has the
	// expected length of 20.  That's "YYYY-MM-DDTHH:MM:SSZ" with no nanoseconds on it.
	// We don't want to index any nanoseconds.
	if len(key) > 20 {
		// Strip off all the nanoseconds (and the Z) then put the Z back on.
		key = key[:19] + "Z"
	}

	return dateRangeToHeightSearchKeyPrefix + key
}

// Return the blockchain height from the index for the given timestamp.  The trunc method is
// for rounding the time to a multiple of dateRangeInterval seconds past midnight.
func (search *Client) getHeightFromTime(
	timeParam string, truncMethod func(int) int,
) (uint64, error) {
	t, err := time.Parse(time.RFC3339, timeParam)
	if err != nil {
		return 0, err
	}

	// Truncate to appropriate snapshot interval.
	seconds := secondsAfterMidnight(t)
	seconds = truncMethod(seconds)
	t = timeAfterMidnight(t, seconds)

	key := formatDateRangeToHeightSearchKey(t)
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
