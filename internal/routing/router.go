package routing

import "github.com/labstack/echo/v5"

type ChoreTypeHandler interface {
	GetBatch(*echo.Context) error
	Create(ctx *echo.Context) error
}

type RouterDeps struct {
	ChoreTypeHandler ChoreTypeHandler
}

type Router struct {
	choreTypeHandler ChoreTypeHandler
}

func NewRouter(deps *RouterDeps) *Router {
	return &Router{
		choreTypeHandler: deps.ChoreTypeHandler,
	}
}
