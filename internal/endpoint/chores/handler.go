package chores

import (
	"context"

	"github.com/a-h/templ"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	choretemplatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	choreviewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
	confirmationviewmodels "github.com/bpsoos/shiftbell/internal/models/view/confirmation"
	"github.com/labstack/echo/v5"
)

type ChoreTemplateService interface {
	Get(context.Context, int) (*choretemplatemodels.ChoreTemplateDetails, error)
}

type Service interface {
	Browse(
		context.Context,
		*choremodels.BrowseChoresParams,
	) (*choremodels.ChorePage, error)
	Create(
		context.Context,
		*choremodels.CreateChoreParams,
	) (*choremodels.CreateChoreResult, error)
	Complete(
		context.Context,
		*choremodels.CompleteChoreParams,
	) (*choremodels.CompleteChoreResult, error)
	CorrectCompletion(
		context.Context,
		*choremodels.CorrectCompletionParams,
	) (*choremodels.ChoreDetails, error)
	Edit(
		context.Context,
		*choremodels.EditChoreParams,
	) (*choremodels.ChoreDetails, error)
	Delete(context.Context, int) error
	Get(context.Context, int) (*choremodels.ChoreDetails, error)
}

type formFeedback struct {
	Values      map[string]string
	FieldErrors map[string]string
	Error       api.ErrorResponse
	Action      api.Relation
	CancelHref  string
}

type View interface {
	Collection(choreviewmodels.Collection, bool) templ.Component
	CompletionDialog(choreviewmodels.CompletionDialog) templ.Component
	ConfirmationDialog(confirmationviewmodels.Dialog) templ.Component
	Detail(choreviewmodels.Detail, bool) templ.Component
	Creation(choreviewmodels.Creation, bool) templ.Component
	ManualOneOffForm(choreviewmodels.ManualOneOffForm, bool) templ.Component
	TemplateOneOffForm(choreviewmodels.TemplateOneOffForm, bool) templ.Component
	EditForm(choreviewmodels.EditForm, bool) templ.Component
	Error(choreviewmodels.Error, bool) templ.Component
}

type HandlerDeps struct {
	View                 View
	Service              Service
	ChoreTemplateService ChoreTemplateService
}

type Handler struct {
	view                 View
	service              Service
	choreTemplateService ChoreTemplateService
}

func NewHandler(deps *HandlerDeps) *Handler {
	return &Handler{
		view:                 deps.View,
		service:              deps.Service,
		choreTemplateService: deps.ChoreTemplateService,
	}
}

func (h *Handler) Get(ctx *echo.Context) error {
	if !supported(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	return h.get(ctx)
}

func (h *Handler) GetBatch(ctx *echo.Context) error {
	if !supported(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	return h.browse(ctx)
}

func (h *Handler) Patch(ctx *echo.Context) error {
	if hypermedia.Accepts(ctx.Request()) {
		return h.editVendorJSON(ctx)
	}
	ctx.Response().Header().Set(echo.HeaderVary, "Accept, HX-Request")
	if !acceptsHTMXHTML(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	return h.editHTMX(ctx)
}

func (h *Handler) Create(ctx *echo.Context) error {
	if !supported(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	return h.create(ctx)
}

func (h *Handler) New(ctx *echo.Context) error {
	if !supported(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	return h.newChore(ctx)
}

func (h *Handler) ConfirmDeletion(ctx *echo.Context) error {
	ctx.Response().Header().Set(echo.HeaderVary, "Accept, HX-Request")
	if !acceptsHTMXHTML(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	return h.confirmDeletion(ctx)
}

func (h *Handler) Edit(ctx *echo.Context) error {
	ctx.Response().Header().Set(echo.HeaderVary, "Accept, HX-Request")
	if !acceptsHTMXHTML(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	return h.editForm(ctx)
}
