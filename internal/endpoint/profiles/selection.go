package profiles

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/bpsoos/shiftbell/internal/endpoint/binding"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/endpoint/profileselection"
	models "github.com/bpsoos/shiftbell/internal/models/profiles"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/profiles"
	"github.com/labstack/echo/v5"
)

const (
	selectionPath = "/profiles/selection"
)

type selectionRequest struct {
	ProfileId int    `form:"profile_id" json:"profile_id"`
	ReturnTo  string `form:"return_to"  json:"return_to"`
}

func (h *Handler) Selection(ctx *echo.Context) error {
	if !h.uiEnabled {
		return ctx.NoContent(http.StatusNotFound)
	}
	if hypermedia.Negotiate(ctx.Request()) != hypermedia.RepresentationHTML {
		return hypermedia.NotAcceptable(ctx)
	}
	profiles, err := h.service.List(ctx.Request().Context())
	if err != nil {
		return h.internalError(ctx, "list profiles for selection", err)
	}
	returnTo := selectionReturnTo(ctx.QueryParam("return_to"))
	if len(profiles) == 0 {
		return hypermedia.HTMLRedirect(ctx, http.StatusSeeOther, "/setup")
	}
	if len(profiles) == 1 {
		return hypermedia.HTMLRedirect(ctx, http.StatusSeeOther, returnTo)
	}
	return h.renderSelection(ctx, http.StatusOK, profiles, returnTo, "")
}

func (h *Handler) Select(ctx *echo.Context) error {
	if !h.uiEnabled {
		return ctx.NoContent(http.StatusNotFound)
	}
	if hypermedia.Negotiate(ctx.Request()) != hypermedia.RepresentationHTML {
		return hypermedia.NotAcceptable(ctx)
	}
	var request selectionRequest
	bindErr := binding.Bind(ctx, &request)
	profiles, err := h.service.List(ctx.Request().Context())
	if err != nil {
		return h.internalError(ctx, "list profiles for selection", err)
	}
	returnTo := selectionReturnTo(request.ReturnTo)
	if request.ReturnTo == "" {
		if referer, err := url.Parse(
			ctx.Request().Referer(),
		); err == nil &&
			referer.Host == ctx.Request().Host {
			returnTo = selectionReturnTo(referer.RequestURI())
		}
	}
	if bindErr != nil {
		return h.renderSelection(
			ctx,
			http.StatusBadRequest,
			profiles,
			returnTo,
			"The selection could not be read. Try again.",
		)
	}
	for _, profile := range profiles {
		if profile.Id == request.ProfileId {
			ctx.SetCookie(&http.Cookie{
				Name:     profileselection.CookieName,
				Value:    strconv.Itoa(profile.Id),
				Path:     "/",
				MaxAge:   365 * 24 * 60 * 60,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
			})
			return hypermedia.HTMLRedirect(ctx, http.StatusSeeOther, returnTo)
		}
	}
	return h.renderSelection(
		ctx,
		http.StatusUnprocessableEntity,
		profiles,
		returnTo,
		"Choose an available profile.",
	)
}

func (h *Handler) renderSelection(
	ctx *echo.Context,
	status int,
	profiles []models.Profile,
	returnTo, message string,
) error {
	ctx.Response().Header().Set("Cache-Control", "private, no-store")
	return hypermedia.HTML(ctx, status, h.view.Selection(viewmodels.Selection{
		Profiles:          profiles,
		SelectedProfileId: profileselection.SelectedId(ctx.Request()),
		ReturnTo:          returnTo,
		Error:             message,
	}, fullPage(ctx)))
}

func selectionReturnTo(value string) string {
	target, err := url.Parse(value)
	if err != nil || target.IsAbs() || target.Host != "" ||
		!strings.HasPrefix(target.Path, "/") || strings.HasPrefix(target.Path, "//") ||
		strings.ContainsAny(
			target.Path,
			"\\\r\n",
		) || target.Path == selectionPath || target.Path == "/setup" {
		return "/chores"
	}
	return target.RequestURI()
}
