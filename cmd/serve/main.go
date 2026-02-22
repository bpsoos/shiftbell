package main

import (
	"log"
	"os"

	choretypesendpoint "github.com/bpsoos/shiftbell/internal/endpoint/choretypes"
	"github.com/bpsoos/shiftbell/internal/persistence"
	"github.com/bpsoos/shiftbell/internal/routing"
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
	chorePersister := persistence.NewChoreTypePersister(&persistence.PersisterDeps{
		Db: db,
	})

	choreTypesTemplater := choretypesview.NewTemplater()
	choreTypesHandler := choretypesendpoint.NewHandler(&choretypesendpoint.HandlerDeps{
		Templater:          choreTypesTemplater,
		ChoreTypePersister: chorePersister,
	})
	router := routing.NewRouter(&routing.RouterDeps{
		ChoreTypeHandler: choreTypesHandler,
	})
	e := echo.New()
	e.Use(middleware.RequestLogger())

	router.Setup(e)

	if err := e.Start("0.0.0.0:80"); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}
