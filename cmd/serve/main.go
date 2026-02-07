package main

import (
	"github.com/bpsoos/shiftbell/internal/endpoint"
	"github.com/bpsoos/shiftbell/internal/routing"
	"github.com/bpsoos/shiftbell/internal/view/home"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	homeTemplater := home.NewTemplater()
	homeHandler := endpoint.NewHandler(&endpoint.HandlerDeps{
		Templater: homeTemplater,
	})
	router := routing.NewRouter(&routing.RouterDeps{
		HomeHandler: homeHandler,
	})
	e := echo.New()
	e.Use(middleware.RequestLogger())

	router.Setup(e)

	if err := e.Start("0.0.0.0:80"); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}

