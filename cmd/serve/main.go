package main

import (
	"context"
	"os"

	"github.com/bpsoos/shiftbell/internal/appcfg"
	"github.com/bpsoos/shiftbell/internal/database"
	choresendpoint "github.com/bpsoos/shiftbell/internal/endpoint/chores"
	choretypesendpoint "github.com/bpsoos/shiftbell/internal/endpoint/choretypes"
	"github.com/bpsoos/shiftbell/internal/logging"
	chorespersistence "github.com/bpsoos/shiftbell/internal/persistence/chores"
	choretypespersistence "github.com/bpsoos/shiftbell/internal/persistence/choretypes"
	"github.com/bpsoos/shiftbell/internal/routing"
	choresview "github.com/bpsoos/shiftbell/internal/view/chores"
	choretypesview "github.com/bpsoos/shiftbell/internal/view/choretypes"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	_ "modernc.org/sqlite"
)

func main() {
	os.Exit(execute())
}

func execute() int {
	logger := logging.Default()
	appConfig, err := appcfg.Load(context.Background())
	if err != nil {
		logger.Error("loading app config", "err", err)
		return 1
	}

	db, err := sqlx.Connect("sqlite", database.SQLiteDSN(appConfig.DatabaseFilepath))
	if err != nil {
		logger.Error("could not connect to db", "err", err)
		return 1
	}
	defer db.Close()
	choreTypePersister := choretypespersistence.NewChoreTypePersister(&choretypespersistence.PersisterDeps{
		Db: db,
	})

	choreTypesTemplater := choretypesview.NewTemplater()
	choreTypesHandler := choretypesendpoint.NewHandler(&choretypesendpoint.HandlerDeps{
		Templater:          choreTypesTemplater,
		ChoreTypePersister: choreTypePersister,
	})

	choresPersister := chorespersistence.NewPersister(&chorespersistence.PersisterDeps{Db: db})
	choresTemplater := choresview.NewTemplater()
	choresHandler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
		Templater:          choresTemplater,
		Persister:          choresPersister,
		ChoreTypePersister: choreTypePersister,
	})

	router := routing.NewRouter(&routing.RouterDeps{
		ChoreHandler:     choresHandler,
		ChoreTypeHandler: choreTypesHandler,
	})
	e := echo.New()
	e.Use(middleware.RequestLogger())

	router.Setup(e)

	if err := e.Start("0.0.0.0:80"); err != nil {
		e.Logger.Error("fatal error", "error", err)
		return 1
	}
	return 0
}
