package main

import (
	"context"
	"os"
	"time"

	"github.com/bpsoos/shiftbell/internal/app"
	"github.com/bpsoos/shiftbell/internal/appcfg"
	"github.com/bpsoos/shiftbell/internal/demodata"
	"github.com/bpsoos/shiftbell/internal/logging"
	chorespersistence "github.com/bpsoos/shiftbell/internal/persistence/chores"
	populatepersistence "github.com/bpsoos/shiftbell/internal/persistence/populate"
	populateservice "github.com/bpsoos/shiftbell/internal/service/populate"
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
	data, err := demodata.Load(os.Getenv("DEMO_DATA_FILEPATH"))
	if err != nil {
		logger.Error("loading demo data", "err", err)
		return 1
	}
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
	service := populateservice.NewService(&populateservice.Deps{
		DataPersister: populatepersistence.NewPersister(
			&populatepersistence.PersisterDeps{
				Db: db,
			},
		),
		ChorePersister: chorespersistence.NewPersister(&chorespersistence.PersisterDeps{
			Db: db,
		}),
		Now: time.Now,
	}, &populateservice.Config{AppTimezone: cfg.AppTimezone})
	if err := service.Populate(context.Background(), data); err != nil {
		logger.Error("populating demo database", "err", err)
		return 1
	}
	return 0
}
