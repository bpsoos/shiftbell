package main

import (
	"os"

	"github.com/bpsoos/shiftbell/internal/app"
	"github.com/bpsoos/shiftbell/internal/appcfg"
	"github.com/bpsoos/shiftbell/internal/logging"
)

func main() {
	os.Exit(execute())
}

func execute() int {
	cfg, err := appcfg.Load()
	if err != nil {
		logging.Default().Error("loading app config", "err", err)
		return 1
	}
	logger := logging.Configure(logging.Config{Handler: cfg.LogHandler})
	db, err := app.OpenDatabase(cfg.DatabaseFilepath)
	if err != nil {
		logger.Error("opening database", "err", err)
		return 1
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("closing database", "err", err)
		}
	}()
	if err := app.Serve(cfg, db); err != nil {
		logger.Error("serving application", "err", err)
		return 1
	}
	return 0
}
