package search

// Types common to all blockchains.

import (
	"errors"
	"fmt"
	"strings"
	"strconv"
)

// DateRangeRequest is used for passing date range query terms over endpoints.
type DateRangeRequest struct {
	FirstTimestamp string
	LastTimestamp  string
}

// DateRangeResult is used for returning search results for the date range endpoint.
type DateRangeResult struct {
	FirstHeight uint64
	LastHeight  uint64
}

// Marshal the request.
func (request *DateRangeRequest) Marshal() string {
	return fmt.Sprintf("%s %s", request.FirstTimestamp, request.LastTimestamp)
}

// Unmarshal the request.
func (request *DateRangeRequest) Unmarshal(requestString string) error {
	separator := strings.Index(requestString, " ")
	if separator < 0 {
		return errors.New("Invalid request string")
	}

	firstTimestamp := requestString[:separator]
	lastTimestamp := requestString[separator+1:]

	request.FirstTimestamp = firstTimestamp
	request.LastTimestamp = lastTimestamp

	return nil
}

// Marshal the result.
func (result *DateRangeResult) Marshal() string {
	return fmt.Sprintf("%d %d", result.FirstHeight, result.LastHeight)
}

// Unmarshal the result.
func (result *DateRangeResult) Unmarshal(resultString string) error {
	separator := strings.Index(resultString, " ")
	if separator < 0 {
		return errors.New("Invalid result string")
	}

	firstHeight, err := strconv.ParseUint(resultString[:separator], 10, 64)
	if err != nil {
		return err
	}

	lastHeight, err := strconv.ParseUint(resultString[separator+1:], 10, 64)
	if err != nil {
		return err
	}

	result.FirstHeight = firstHeight
	result.LastHeight = lastHeight

	return nil
}
