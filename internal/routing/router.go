package routing

import "github.com/labstack/echo/v5"

type HomeHandler interface {
	Home(*echo.Context) error
	GetChoreBatch(ctx *echo.Context) error
	CreateChore(ctx *echo.Context) error
	ViewSettings(ctx *echo.Context) error
}

type RouterDeps struct {
	HomeHandler HomeHandler
}

type Router struct {
	homeHandler HomeHandler
}

func NewRouter(deps *RouterDeps) *Router {
	return &Router{
		homeHandler: deps.HomeHandler,
	}
}
