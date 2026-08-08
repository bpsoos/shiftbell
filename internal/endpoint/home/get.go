package home

import (
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	"github.com/labstack/echo/v5"
)

type response struct {
	Links map[string]api.Link `json:"_links"`
}

func (h *Handler) Get(ctx *echo.Context) error {
	switch hypermedia.Negotiate(ctx.Request()) {
	case hypermedia.RepresentationJSON:
		return hypermedia.JSON(ctx, http.StatusOK, response{
			Links: map[string]api.Link{
				"self":            {Href: "/"},
				"chores":          {Href: "/chores"},
				"chore_templates": {Href: "/chore-templates"},
			},
		})
	case hypermedia.RepresentationHTML:
		return hypermedia.HTMLRedirect(ctx, http.StatusSeeOther, "/chores")
	default:
		return hypermedia.NotAcceptable(ctx)
	}
}
