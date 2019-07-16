package app

import (
	"os"

	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	log "github.com/sirupsen/logrus"
)

// NewLogger creates a new logger with default configuration,
// some of which can be overridden from environment variables.
// Callers should set up node_id and bin fields on the returned logger.
func NewLogger() log.FieldLogger {
	logger := log.New()
	logger.Out = os.Stderr

	var formatter log.Formatter
	switch os.Getenv("LOG_FORMAT") {
	case "json", "":
		formatter = &log.JSONFormatter{}
	case "text", "plain":
		formatter = &log.TextFormatter{}
	default:
		formatter = &log.JSONFormatter{}
	}
	logger.Formatter = formatter

	var level log.Level
	switch os.Getenv("LOG_LEVEL") {
	case "info", "":
		level = log.InfoLevel
	case "debug":
		level = log.DebugLevel
	case "warn", "warning":
		level = log.WarnLevel
	case "err", "error":
		level = log.ErrorLevel
	default:
		level = log.InfoLevel
	}
	logger.Level = level

	return logger
}

// GetLogger returns the application logger
func (app *App) GetLogger() log.FieldLogger {
	return app.logger
}

// SetLogger sets the logger to be used by this app.
func (app *App) SetLogger(logger log.FieldLogger) {
	app.logger = logger
}

// DecoratedLogger returns a logger decorated with standard app data
func (app *App) DecoratedLogger() *log.Entry {
	return app.logger.WithFields(log.Fields{
		"height":        app.Height(),
		"hash":          app.HashStr(),
		"app.blockTime": app.blockTime.String(),
	})
}

// DecoratedTxLogger returns a logger decorated with the tx hash
func (app *App) DecoratedTxLogger(tx metatx.Transactable) *log.Entry {
	return app.DecoratedLogger().WithField("tx hash", metatx.Hash(tx))
}

// LogState emits a log message detailing the current app state
func (app *App) LogState() {
	app.DecoratedLogger().Info("LogState")
}

// logRequest emits a log message on request receipt
//
// It also returns a decorated logger for request-internal logging.
func (app *App) logRequestOptHt(method string, showHeight bool, logger log.FieldLogger) log.FieldLogger {
	if logger == nil {
		logger = app.GetLogger()
	}
	decoratedLogger := logger.WithField("method", method)
	if showHeight {
		decoratedLogger = decoratedLogger.WithField("height", app.Height())
		decoratedLogger = decoratedLogger.WithField("hash", app.HashStr())
	}
	decoratedLogger.Info("received request")

	return decoratedLogger
}

func (app *App) logRequest(m string, logger log.FieldLogger) log.FieldLogger {
	return app.logRequestOptHt(m, true, logger)
}

func (app *App) logRequestBare(m string, logger log.FieldLogger) log.FieldLogger {
	return app.logRequestOptHt(m, false, logger)
}
