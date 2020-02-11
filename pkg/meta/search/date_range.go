package search

// ----- ---- --- -- -
// Copyright 2019, 2020 The Axiom Foundation. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----


// Blockchain-independent implementation for date range indexing and searching.

import (
	"fmt"
	"strconv"
	"strings"

	math "github.com/oneiro-ndev/ndaumath/pkg/types"
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
const dateRangeToHeightSearchKeyPrefix = "date.range:height:"

// Return the number of seconds past midnight of the given time, ignoring nanoseconds.
func secondsAfterMidnight(t math.Timestamp) int {
	return int(t%math.Day) / math.Second
}

// Return the given time, truncated to midnight, then with the given seconds added to it.
func timeAfterMidnight(t math.Timestamp, seconds int) math.Timestamp {
	trunc := t / math.Day
	return (trunc * math.Day) + math.Timestamp(seconds*math.Second)
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
func truncTime(t math.Timestamp, truncMethod func(int) int) math.Timestamp {
	seconds := secondsAfterMidnight(t)
	seconds = truncMethod(seconds)
	return timeAfterMidnight(t, seconds)
}

// Format the given time into a date key that we index and search on for date range queries.
// The time should already be trunctated to a multiple of dateRangeInterval seconds past midnight.
func formatDateRangeToHeightSearchKey(truncatedTime math.Timestamp) string {
	key := truncatedTime.String()

	return dateRangeToHeightSearchKeyPrefix + key
}

// Return the blockchain height from the index for the given timestamp.
// truncMethod is for rounding the time to a multiple of dateRangeInterval seconds past midnight.
func (search *Client) getHeightFromTime(
	timeParam string, truncMethod func(int) int,
) (uint64, error) {
	t, err := math.ParseTimestamp(timeParam)
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
func getPrevAndNextTimes(t math.Timestamp) (prevTime, nextTime math.Timestamp) {
	prevTime = truncTime(t, floorSeconds)
	nextTime = truncTime(prevTime.Add(math.Duration(math.Second)), ceilSeconds)
	return
}

// IndexDateToHeight will index all necessary date-to-height keys back in time to the latest one
// we've indexed, using the given date and height.  Typically this function will only need to do
// work once every dateRangeInterval seconds.  But if there are long periods of block inactivity,
// this function will fill in all missing date-to-height keys up to the given block time.
// The given block height must be > 0, which is guaranteed if it comes from Tendermint.
func (search *Client) IndexDateToHeight(
	blockTime math.Timestamp, blockHeight uint64,
) (updateCount int, insertCount int, err error) {
	updateCount = 0
	insertCount = 0

	// Ignore invalid block times.
	if blockTime < 0 {
		return updateCount, insertCount, nil
	}

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

	// These next steps ensure that we've initialized the date range index.
	// The earliest time also prevents any chance of an infinite loop later in this function.
	var earliestTime math.Timestamp
	earliestKey, err := search.Get(dateRangeToHeightSearchKeyPrefix)
	if err != nil {
		return updateCount, insertCount, err
	}
	if strings.Index(earliestKey, dateRangeToHeightSearchKeyPrefix) != 0 {
		// The earliest key doesn't exist yet.  So the block time is going to be our genesis.
		// Typically this happens on the first block we index, so it's what we want.  If it's not
		// the first block, it means we've upgraded the code without starting a fresh blockchain.
		// In that case, date range queries before this block will return empty results.  Some
		// blockchains don't index timestamps in their initial indexers, so we didn't bump the
		// index version to force a wipe and reindex.  We do the best with what we've got here.
		earliestTime = prevTime
	} else {
		// Grab the earliest time out of the index.
		earliestTimestamp := earliestKey[len(dateRangeToHeightSearchKeyPrefix):]
		earliestTime, err = math.ParseTimestamp(earliestTimestamp)
		if err != nil {
			return updateCount, insertCount, err
		}
		earliestValue, err := search.Get(earliestKey)
		if err != nil {
			return updateCount, insertCount, err
		}
		if earliestValue != nextHeightFlag {
			return updateCount, insertCount, fmt.Errorf(
				"Unexpected earliest value '%s' in key %s", earliestValue, earliestKey)
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
	if blockTime != prevTime {
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

		if prevTime <= earliestTime {
			// Prevent infinite loop in case something went wrong in the index.
			break
		}

		prevTime = prevTime.Add(math.Duration(-dateRangeInterval * math.Second))
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
