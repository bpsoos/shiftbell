package routing

import "github.com/labstack/echo/v5"

type ChoreTemplateHandler interface {
	GetBatch(*echo.Context) error
	Create(ctx *echo.Context) error
	Delete(ctx *echo.Context) error
}

type ChoreHandler interface {
	Get(*echo.Context) error
	GetBatch(*echo.Context) error
	Patch(*echo.Context) error
	New(*echo.Context) error
	Create(*echo.Context) error
}

type RouterDeps struct {
	ChoreTemplateHandler ChoreTemplateHandler
	ChoreHandler         ChoreHandler
}

type Router struct {
	choreTemplateHandler ChoreTemplateHandler
	choreHandler         ChoreHandler
}

func NewRouter(deps *RouterDeps) *Router {
	return &Router{
		choreTemplateHandler: deps.ChoreTemplateHandler,
		choreHandler:         deps.ChoreHandler,
	}
}
