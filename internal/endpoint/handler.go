package endpoint

import "github.com/labstack/echo/v5"

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Home(ctx *echo.Context) error {
	return nil
}
