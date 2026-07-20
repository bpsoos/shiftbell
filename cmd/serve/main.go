package main

import (
	"os"

	"github.com/bpsoos/shiftbell/internal/appcfg"
	"github.com/bpsoos/shiftbell/internal/database"
	choresendpoint "github.com/bpsoos/shiftbell/internal/endpoint/chores"
	choretemplatesendpoint "github.com/bpsoos/shiftbell/internal/endpoint/choretemplates"
	"github.com/bpsoos/shiftbell/internal/logging"
	chorespersistence "github.com/bpsoos/shiftbell/internal/persistence/chores"
	choretemplatespersistence "github.com/bpsoos/shiftbell/internal/persistence/choretypes"
	"github.com/bpsoos/shiftbell/internal/routing"
	choresview "github.com/bpsoos/shiftbell/internal/view/chores"
	choretemplatesview "github.com/bpsoos/shiftbell/internal/view/choretemplates"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	_ "modernc.org/sqlite"
)

func main() {
	os.Exit(execute())
}

func execute() int {
	appConfig, err := appcfg.Load()
	if err != nil {
		logging.Default().Error("loading app config", "err", err)
		return 1
	}
	logger := logging.Configure(logging.Config{Handler: appConfig.LogHandler})

	db, err := sqlx.Connect("sqlite", database.SQLiteDSN(appConfig.DatabaseFilepath))
	if err != nil {
		logger.Error("could not connect to db", "err", err)
		return 1
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	choreTemplatePersister := choretemplatespersistence.NewChoreTypePersister(&choretemplatespersistence.PersisterDeps{
		Db: db,
	})

	choreTemplatesTemplater := choretemplatesview.NewTemplater()
	choreTemplatesHandler := choretemplatesendpoint.NewHandler(&choretemplatesendpoint.HandlerDeps{
		Templater:              choreTemplatesTemplater,
		ChoreTemplatePersister: choreTemplatePersister,
	})

	choresPersister := chorespersistence.NewPersister(&chorespersistence.PersisterDeps{Db: db})
	choresTemplater := choresview.NewTemplater()
	choresHandler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
		Templater:              choresTemplater,
		Persister:              choresPersister,
		ChoreTemplatePersister: choreTemplatePersister,
	})

	router := routing.NewRouter(&routing.RouterDeps{
		ChoreHandler:         choresHandler,
		ChoreTemplateHandler: choreTemplatesHandler,
	})
	e := echo.New()
	e.Logger = logger
	e.Use(middleware.RequestLogger())

	router.Setup(e)

	if err := e.Start("0.0.0.0:80"); err != nil {
		logging.Default().Error("fatal error", "error", err)
		return 1
	}
	return 0
}
