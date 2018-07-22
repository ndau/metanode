package app

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"

	libhoney "github.com/honeycombio/libhoney-go"
	"github.com/sirupsen/logrus"
)

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

// There are two things we should only do once -- one is initialize the libhoney library,
// and the other is registration of the logrus hook. Consequently, we need two instances
// of a sync.Once.
var initHoneycombOnce sync.Once
var registerLogglyOnce sync.Once

// initHoneycomb initializes honeycomb by calling the init function only once.
func initHoneycomb() error {
	var err error
	initHoneycombOnce.Do(func() {
		writeKey := os.Getenv("HONEYCOMB_KEY")
		datasetName := os.Getenv("HONEYCOMB_DATASET")
		cfg := libhoney.Config{
			WriteKey: writeKey,
			Dataset:  datasetName,
		}
		err = libhoney.Init(cfg)
		if err != nil {
			return
		}
		_, err = libhoney.VerifyWriteKey(cfg)
	})
	return err
}

// newHoneycombHook returns a new Honeycomb.io logrus hook
func newHoneycombHook() (logrus.Hook, error) {
	err := initHoneycomb()
	if err != nil {
		return nil, err
	}
	return &honeycombHook{}, nil
}

// SetupHoneycomb sets up a logrus logger to send its data to honeycomb instead of
// sending it to stdout.
func SetupHoneycomb(logger *logrus.Logger) *logrus.Logger {
	honeycombLoggingHook, err := newHoneycombHook()
	if err != nil {
		logger.Warn(err)
		logger.Warn("Honeycomb failed to initialize properly - did you set HONEYCOMB_KEY and HONEYCOMB_DATASET?")
	}

	registerLogglyOnce.Do(func() {
		logger.Hooks.Add(honeycombLoggingHook)
		logger.Out = ioutil.Discard
	})

	logger.WithFields(logrus.Fields{
		"bee_stings": rand.Int31n(10),
	}).Info("Ouch!")
	return logger
}

type honeycombWriter struct{}

// Write implements io.Writer for honeycombWriter; it assumes that b is a JSON blob
// and unmarshals it into an interface{}, then simply sends that as a new event
// to honeycomb.
func (h *honeycombWriter) Write(b []byte) (int, error) {
	var v interface{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return 0, err
	}
	evt := libhoney.NewBuilder().NewEvent()
	err = evt.Add(v)
	if err != nil {
		return 0, err
	}
	err = evt.Send()
	if err != nil {
		return 0, err
	}

	return len(b), nil
}

var _ io.Writer = (*honeycombWriter)(nil)

// NewHoneycombWriter constructs a writer that assumes its input is JSON and
// sends it to Honeycomb.
func NewHoneycombWriter() (io.Writer, error) {
	err := initHoneycomb()
	if err != nil {
		return nil, err
	}
	return &honeycombWriter{}, nil
}
