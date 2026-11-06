package profiles

import (
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/labstack/echo/v5"
)

func (h *Handler) RequireSetup(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx *echo.Context) error {
		if !h.uiEnabled ||
			hypermedia.Negotiate(ctx.Request()) != hypermedia.RepresentationHTML {
			return next(ctx)
		}
		if h.setupRequired.Load() {
			return hypermedia.HTMLRedirect(ctx, http.StatusSeeOther, "/setup")
		}
		return next(ctx)
	}
}
