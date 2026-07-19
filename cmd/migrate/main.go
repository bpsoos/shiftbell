package main

import (
	"os"

	"github.com/bpsoos/shiftbell/internal/appcfg"
	"github.com/bpsoos/shiftbell/internal/logging"
	"github.com/bpsoos/shiftbell/internal/migrations"
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

	err = migrations.NewMigrator().Migrate(cfg.DatabaseFilepath)
	if err != nil {
		logger.Error("error running migrations", "err", err)
		return 1
	}
	return 0
}
