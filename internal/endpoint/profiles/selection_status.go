package profiles

import (
	"net/http"
	"net/url"

	"github.com/a-h/templ"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/endpoint/profileselection"
	"github.com/bpsoos/shiftbell/internal/logging"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/profiles"
	"github.com/labstack/echo/v5"
)

func (h *Handler) SelectionStatus(ctx *echo.Context) error {
	if !h.uiEnabled {
		return ctx.NoContent(http.StatusNotFound)
	}
	if hypermedia.Negotiate(ctx.Request()) != hypermedia.RepresentationHTML {
		return hypermedia.NotAcceptable(ctx)
	}
	ctx.Response().Header().Set("Cache-Control", "private, no-store")
	profiles, err := h.service.List(ctx.Request().Context())
	if err != nil {
		logging.Default().Error("resolve selected profile", "err", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}
	if len(profiles) == 0 {
		return hypermedia.HTMLRedirect(ctx, http.StatusSeeOther, "/setup")
	}
	if len(profiles) == 1 {
		return hypermedia.HTML(ctx, http.StatusOK, templ.NopComponent)
	}
	selected := profileselection.SelectedId(ctx.Request())
	for _, profile := range profiles {
		if profile.Id == selected {
			return hypermedia.HTML(
				ctx,
				http.StatusOK,
				h.view.SelectedProfile(viewmodels.Navigation{
					Profile: profile, Profiles: profiles,
				}),
			)
		}
	}
	switchHref := selectionPath + "?" + url.Values{
		"return_to": {selectionDestination(ctx)},
	}.Encode()
	return hypermedia.HTMLRedirect(ctx, http.StatusSeeOther, switchHref)
}

func selectionDestination(ctx *echo.Context) string {
	current, err := url.Parse(ctx.Request().Header.Get("HX-Current-URL"))
	if err == nil && current.Host == ctx.Request().Host {
		return selectionReturnTo(current.RequestURI())
	}
	return "/chores"
}
