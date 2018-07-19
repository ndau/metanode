package app

import (
	"io/ioutil"
	"math/rand"
	"os"
	"sync"

	libhoney "github.com/honeycombio/libhoney-go"
	"github.com/sirupsen/logrus"
)

// one time logger registration
var oneTime sync.Once

////////////////////////////////////////////////////////////////////////////////
// Honeycomb.io Logrus hook
////////////////////////////////////////////////////////////////////////////////
type honeycombHook struct {
}

func (hook *honeycombHook) Fire(entry *logrus.Entry) error {
	eventBuilder := libhoney.NewBuilder()
	honeycombEvent := eventBuilder.NewEvent()
	for eachKey, eachValue := range entry.Data {
		honeycombEvent.AddField(eachKey, eachValue)
	}
	honeycombEvent.AddField("ts", entry.Time)
	honeycombEvent.Send()
	return nil
}

func (hook *honeycombHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.InfoLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}

////////////////////////////////////////////////////////////////////////////////
// Return a new Honeycomb.io logrus hook
////////////////////////////////////////////////////////////////////////////////
func newHoneycombHook(writeKey string, datasetName string) (logrus.Hook, error) {
	cfg := libhoney.Config{
		WriteKey: writeKey,
		Dataset:  datasetName,
	}
	err := libhoney.Init(cfg)
	if err != nil {
		return nil, err
	}
	_, err = libhoney.VerifyWriteKey(cfg)
	if err != nil {
		return nil, err
	}
	return &honeycombHook{}, nil
}

// SetupHoneycomb sets up a logrus logger to send its data to honeycomb instead of
// sending it to stdout.
func SetupHoneycomb(logger *logrus.Logger) *logrus.Logger {
	// Lazily register the logrus logging hook
	oneTime.Do(func() {
		key := os.Getenv("HONEYCOMB_KEY")
		dataset := os.Getenv("HONEYCOMB_DATASET")
		honeycombLoggingHook, err := newHoneycombHook(key, dataset)
		if err == nil {
			logger.Hooks.Add(honeycombLoggingHook)
			logger.Out = ioutil.Discard
			logger.Warn("Honeycomb failed to initialize properly - did you set HONEYCOMB_KEY and HONEYCOMB_DATASET?")
		}
	})

	logger.WithFields(logrus.Fields{
		"bee_stings": rand.Int31n(10),
	}).Info("Ouch!")
	return logger
}
