package main

import (
	"context"
	"os"

	"github.com/bpsoos/shiftbell/internal/appcfg"
	"github.com/bpsoos/shiftbell/internal/logging"
	"github.com/bpsoos/shiftbell/internal/migrations"
)

func main() {
	os.Exit(execute())
}

func execute() int {
	logger := logging.Default()
	cfg, err := appcfg.Load(context.Background())
	if err != nil {
		logger.Error("loading app config", "err", err)
		return 1
	}

	err = migrations.NewMigrator().Migrate(cfg.DatabaseFilepath)
	if err != nil {
		logger.Error("error running migrations", "err", err)
		return 1
	}
	return 0
}
