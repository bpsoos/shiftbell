package main

import (
	"log"
	"os"

	choresendpoint "github.com/bpsoos/shiftbell/internal/endpoint/chores"
	choretypesendpoint "github.com/bpsoos/shiftbell/internal/endpoint/choretypes"
	chorespersistence "github.com/bpsoos/shiftbell/internal/persistence/chores"
	choretypespersistence "github.com/bpsoos/shiftbell/internal/persistence/choretypes"
	"github.com/bpsoos/shiftbell/internal/routing"
	choresview "github.com/bpsoos/shiftbell/internal/view/chores"
	choretypesview "github.com/bpsoos/shiftbell/internal/view/choretypes"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("could not connect to db", err)
	}
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
		e.Logger.Error("failed to start server", "error", err)
	}
}
