package home

import (
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	"github.com/labstack/echo/v5"
)

type response struct {
	Links api.Relations `json:"_links"`
}

func (h *Handler) Get(ctx *echo.Context) error {
	switch hypermedia.Negotiate(ctx.Request()) {
	case hypermedia.RepresentationJSON:
		return hypermedia.JSON(ctx, http.StatusOK, response{
			Links: api.Relations{
				{Rel: "self", Href: "/"},
				{Rel: "chores", Href: "/chores"},
				{Rel: "chore_templates", Href: "/chore-templates"},
			},
		})
	case hypermedia.RepresentationHTML:
		return hypermedia.HTMLRedirect(ctx, http.StatusSeeOther, "/chores")
	default:
		return hypermedia.NotAcceptable(ctx)
	}
}
