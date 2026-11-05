package profiles

import (
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/models/api"
	apimodels "github.com/bpsoos/shiftbell/internal/models/api/profiles"
	"github.com/labstack/echo/v5"
)

func (h *Handler) Setup(ctx *echo.Context) error {
	return h.create(ctx, true)
}

func (h *Handler) SetupStatus(ctx *echo.Context) error {
	if !h.supported(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	required, err := h.service.SetupRequired(ctx.Request().Context())
	if err != nil {
		return h.internalError(ctx, "check profile setup", err)
	}
	if hypermedia.Negotiate(ctx.Request()) == hypermedia.RepresentationHTML {
		if !required {
			return hypermedia.HTMLRedirect(ctx, http.StatusSeeOther, "/chores")
		}
		return hypermedia.HTML(
			ctx,
			http.StatusOK,
			h.view.Form(newForm("", true), fullPage(ctx)),
		)
	}
	actions := api.Relations{}
	if required {
		actions = append(actions, api.Relation{Rel: "create", Href: "/setup"})
	}
	return hypermedia.JSON(ctx, http.StatusOK, apimodels.Setup{
		Required: required,
		Links: api.Relations{
			{Rel: "self", Href: "/setup"},
			{Rel: "profiles", Href: "/profiles"},
			{Rel: "chores", Href: "/chores"},
		},
		Actions: actions,
	})
}
