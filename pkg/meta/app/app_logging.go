package app

import (
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/o11y/pkg/honeycomb"
	log "github.com/sirupsen/logrus"
)

// GetLogger returns the application logger
func (app *App) GetLogger() log.FieldLogger {
	return app.logger
}

// SetLogger sets the logger to be used by this app.
// It has the side effect of setting up Honeycomb if it's possible to do so.
func (app *App) SetLogger(logger log.FieldLogger) {
	switch l := logger.(type) {
	case *log.Logger:
		app.logger = honeycomb.Setup(l)
		app.logger = l
	case *log.Entry:
		l.Logger = honeycomb.Setup(l.Logger)
		app.logger = l
	default:
		logger.Warnf("Logger was %T, so can't set up Honeycomb.", logger)
		app.logger = logger
	}
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
