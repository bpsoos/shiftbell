package profiles

import (
	"context"
	"sync/atomic"

	"github.com/a-h/templ"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/models/api"
	apimodels "github.com/bpsoos/shiftbell/internal/models/api/profiles"
	models "github.com/bpsoos/shiftbell/internal/models/profiles"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/profiles"
	"github.com/labstack/echo/v5"
)

type Service interface {
	Create(context.Context, *models.CreateParams) (*models.Profile, error)
	List(context.Context) ([]models.Profile, error)
	Setup(context.Context, *models.CreateParams) (*models.Profile, error)
}

type View interface {
	SelectedProfile(viewmodels.Navigation) templ.Component
	Selection(viewmodels.Selection, bool) templ.Component
	Collection(viewmodels.Collection, bool) templ.Component
	Form(viewmodels.Form, bool) templ.Component
	Error(viewmodels.Error, bool) templ.Component
}

type HandlerDeps struct {
	Service Service
	View    View
}

type HandlerConfig struct {
	SetupRequired bool
	UIEnabled     bool
}

type Handler struct {
	service       Service
	view          View
	setupRequired atomic.Bool
	uiEnabled     bool
}

func NewHandler(deps *HandlerDeps, config *HandlerConfig) *Handler {
	handler := &Handler{
		service:   deps.Service,
		view:      deps.View,
		uiEnabled: config.UIEnabled,
	}
	handler.setupRequired.Store(config.SetupRequired)
	return handler
}

func (h *Handler) supported(ctx *echo.Context) bool {
	representation := hypermedia.Negotiate(ctx.Request())
	return representation == hypermedia.RepresentationJSON ||
		(representation == hypermedia.RepresentationHTML && h.view != nil)
}

func fullPage(ctx *echo.Context) bool {
	return ctx.Request().Header.Get("HX-Request") != "true"
}

func newResponse(profile *models.Profile) apimodels.Response {
	return apimodels.Response{Id: profile.Id, Name: profile.Name, Links: collectionLink()}
}

func collectionLink() api.Relations {
	return api.Relations{{Rel: "collection", Href: "/profiles"}}
}
