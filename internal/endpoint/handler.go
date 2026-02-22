package endpoint

import (
	"github.com/labstack/echo/v5"
)

type Templater interface {
}

type Persister interface {
}

type HandlerDeps struct {
	Templater Templater
	Persister Persister
}

type Handler struct {
	templater Templater
	persister Persister
}

func NewHandler(deps *HandlerDeps) *Handler {
	return &Handler{
		templater: deps.Templater,
		persister: deps.Persister,
	}
}

func (h *Handler) Home(ctx *echo.Context) error {
	return nil
}
