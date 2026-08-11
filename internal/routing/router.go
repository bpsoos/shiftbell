package routing

import "github.com/labstack/echo/v5"

type HomeHandler interface {
	Get(*echo.Context) error
}

type ChoreTemplateHandler interface {
	Browse(*echo.Context) error
	Create(*echo.Context) error
	Get(*echo.Context) error
	Edit(*echo.Context) error
	Deactivate(*echo.Context) error
}

type ChoreHandler interface {
	Get(*echo.Context) error
	GetBatch(*echo.Context) error
	Patch(*echo.Context) error
	New(*echo.Context) error
	Create(*echo.Context) error
	ConfirmCompletion(*echo.Context) error
	Complete(*echo.Context) error
	CorrectCompletion(*echo.Context) error
	Delete(*echo.Context) error
}

type RouterDeps struct {
	HomeHandler          HomeHandler
	ChoreTemplateHandler ChoreTemplateHandler
	ChoreHandler         ChoreHandler
}

type Router struct {
	homeHandler          HomeHandler
	choreTemplateHandler ChoreTemplateHandler
	choreHandler         ChoreHandler
}

func NewRouter(deps *RouterDeps) *Router {
	return &Router{
		homeHandler:          deps.HomeHandler,
		choreTemplateHandler: deps.ChoreTemplateHandler,
		choreHandler:         deps.ChoreHandler,
	}
}
