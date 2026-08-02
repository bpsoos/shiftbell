package main

import (
	"os"

	"github.com/bpsoos/shiftbell/internal/appcfg"
	"github.com/bpsoos/shiftbell/internal/database"
	choresendpoint "github.com/bpsoos/shiftbell/internal/endpoint/chores"
	choretemplatesendpoint "github.com/bpsoos/shiftbell/internal/endpoint/choretemplates"
	"github.com/bpsoos/shiftbell/internal/endpoint/home"
	"github.com/bpsoos/shiftbell/internal/logging"
	"github.com/bpsoos/shiftbell/internal/migrations"
	chorespersistence "github.com/bpsoos/shiftbell/internal/persistence/chores"
	choretemplatespersistence "github.com/bpsoos/shiftbell/internal/persistence/choretemplates"
	"github.com/bpsoos/shiftbell/internal/routing"
	choresservice "github.com/bpsoos/shiftbell/internal/service/chores"
	choretemplatesservice "github.com/bpsoos/shiftbell/internal/service/choretemplates"
	"github.com/bpsoos/shiftbell/internal/service/normalization"
	choresview "github.com/bpsoos/shiftbell/internal/view/chores"
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

	if err := migrations.NewMigrator().Migrate(appConfig.DatabaseFilepath); err != nil {
		logger.Error("error running migrations", "err", err)
		return 1
	}

	db, err := sqlx.Connect("sqlite", database.SQLiteDSN(appConfig.DatabaseFilepath))
	if err != nil {
		logger.Error("could not connect to db", "err", err)
		return 1
	}
	defer func() {
		err := db.Close()
		if err != nil {
			logger.Error("error closing db", "err", err)
		}
	}()
	db.SetMaxOpenConns(1)
	choreTemplatePersister := choretemplatespersistence.NewChoreTemplatePersister(
		&choretemplatespersistence.PersisterDeps{
			Db: db,
		},
	)

	normalizer := normalization.New(normalization.Config{
		NameLimit:        200,
		DescriptionLimit: 2000,
		SearchLimit:      200,
	})
	choreTemplateService := choretemplatesservice.NewService(
		&choretemplatesservice.Deps{
			Persister:  choreTemplatePersister,
			Normalizer: normalizer,
		},
		&choretemplatesservice.Config{},
	)
	choreTemplatesHandler := choretemplatesendpoint.NewHandler(
		&choretemplatesendpoint.HandlerDeps{
			Service: choreTemplateService,
		},
	)

	choresPersister := chorespersistence.NewPersister(
		&chorespersistence.PersisterDeps{Db: db},
	)
	choresTemplater := choresview.NewTemplater()
	choresService := choresservice.NewService(
		&choresservice.Deps{
			Persister:  choresPersister,
			Normalizer: normalizer,
		},
		&choresservice.Config{AppTimezone: appConfig.AppTimezone},
	)
	choresHandler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
		Templater:              choresTemplater,
		Persister:              choresPersister,
		Service:                choresService,
		ChoreTemplatePersister: choreTemplatePersister,
	})

	router := routing.NewRouter(&routing.RouterDeps{
		HomeHandler:          home.NewHandler(),
		ChoreHandler:         choresHandler,
		ChoreTemplateHandler: choreTemplatesHandler,
	})
	e := echo.New()
	e.Logger = logger
	e.Use(middleware.RequestLogger())

	err = router.Setup(e)
	if err != nil {
		logger.Error("error setting up routes", "err", err)
		return 1
	}

	if err := e.Start("0.0.0.0:80"); err != nil {
		logging.Default().Error("fatal error", "error", err)
		return 1
	}
	return 0
}
