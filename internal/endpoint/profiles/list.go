package profiles

import (
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/models/api"
	apimodels "github.com/bpsoos/shiftbell/internal/models/api/profiles"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/profiles"
	"github.com/labstack/echo/v5"
)

func (h *Handler) List(ctx *echo.Context) error {
	if !h.supported(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	profiles, err := h.service.List(ctx.Request().Context())
	if err != nil {
		return h.internalError(ctx, "list profiles", err)
	}
	items := make([]apimodels.Response, len(profiles))
	for i := range profiles {
		items[i] = newResponse(&profiles[i])
	}
	if hypermedia.Negotiate(ctx.Request()) == hypermedia.RepresentationHTML {
		return hypermedia.HTML(
			ctx,
			http.StatusOK,
			h.view.Collection(viewmodels.Collection{
				Profiles: items, CreateHref: "/profiles/new",
			}, fullPage(ctx)),
		)
	}
	return hypermedia.JSON(ctx, http.StatusOK, apimodels.Collection{
		Items: items,
		Links: api.Relations{
			{Rel: "self", Href: "/profiles"},
			{Rel: "setup", Href: "/setup"},
		},
		Actions: api.Relations{{Rel: "create", Href: "/profiles"}},
	})
}
