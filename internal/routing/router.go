package routing

import "github.com/labstack/echo/v5"

type ChoreTypeHandler interface {
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
	ChoreTypeHandler ChoreTypeHandler
	ChoreHandler     ChoreHandler
}

type Router struct {
	choreTypeHandler ChoreTypeHandler
	choreHandler     ChoreHandler
}

func NewRouter(deps *RouterDeps) *Router {
	return &Router{
		choreTypeHandler: deps.ChoreTypeHandler,
		choreHandler:     deps.ChoreHandler,
	}
}
