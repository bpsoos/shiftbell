package profiles

import (
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/labstack/echo/v5"
)

func (h *Handler) RequireSetup(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx *echo.Context) error {
		if hypermedia.Negotiate(ctx.Request()) != hypermedia.RepresentationHTML {
			return next(ctx)
		}
		required, err := h.service.SetupRequired(ctx.Request().Context())
		if err != nil {
			return h.internalError(ctx, "check profile setup", err)
		}
		if required {
			return hypermedia.HTMLRedirect(ctx, http.StatusSeeOther, "/setup")
		}
		return next(ctx)
	}
}
