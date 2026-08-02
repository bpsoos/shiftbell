package home

import (
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/labstack/echo/v5"
)

type response struct {
	Links map[string]hypermedia.Link `json:"_links"`
}

func (h *Handler) Get(ctx *echo.Context) error {
	if !hypermedia.Accepts(ctx.Request()) {
		return ctx.NoContent(http.StatusNotAcceptable)
	}

	return hypermedia.JSON(ctx, http.StatusOK, response{
		Links: map[string]hypermedia.Link{
			"self":            {Href: "/"},
			"chores":          {Href: "/chores"},
			"chore_templates": {Href: "/chore-templates"},
		},
	})
}
