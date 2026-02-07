package endpoint

import (
	"context"
	"io"
	"net/http"

	"github.com/labstack/echo/v5"
)

type Templater interface {
	Home(context.Context, io.Writer) error
}

type HandlerDeps struct {
	Templater Templater
}

type Handler struct {
	templater Templater
}

func NewHandler(deps *HandlerDeps) *Handler {
	return &Handler{
		templater: deps.Templater,
	}
}

func (h *Handler) Home(ctx *echo.Context) error {
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	ctx.Response().WriteHeader(http.StatusOK)

	return h.templater.Home(ctx.Request().Context(), ctx.Response())
}
