package choretemplates

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/binding"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/choretemplates"
	"github.com/labstack/echo/v5"
)

type createRequest struct {
	Name        string `form:"name"        json:"name"`
	Description string `form:"description" json:"description"`
}

func (h *Handler) Create(ctx *echo.Context) error {
	if !supported(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	request, err := parseCreateRequest(ctx)
	if err != nil {
		return h.renderCreateBindingError(ctx, request, err)
	}

	choreTemplate, err := h.service.Create(
		ctx.Request().Context(),
		&models.CreateChoreTemplateParams{
			Name:        request.Name,
			Description: request.Description,
		},
	)
	if err != nil {
		return h.renderCreateFailure(ctx, request, err)
	}

	representation := newRepresentation(choreTemplate)
	ctx.Response().Header().Set(echo.HeaderLocation, representation.Links.Href("self"))
	if acceptsHTML(ctx) {
		ctx.Response().Header().Set("HX-Replace-Url", representation.Links.Href("self"))
	}
	return h.renderDetail(
		ctx,
		http.StatusCreated,
		representation,
		"Chore template created.",
	)
}

func parseCreateRequest(ctx *echo.Context) (createRequest, error) {
	var request createRequest
	if ctx.Request().Header.Get(echo.HeaderContentType) == "" {
		err := json.NewDecoder(ctx.Request().Body).Decode(&request)
		return request, err
	}
	err := binding.Bind(ctx, &request)
	return request, err
}

func (request createRequest) form() viewmodels.CreateForm {
	return viewmodels.CreateForm{
		ActionHref: choreTemplateCollectionHref,
		CancelHref: choreTemplateCollectionHref,
		Name: viewmodels.Field{
			Value: request.Name,
		},
		Description: viewmodels.Field{Value: request.Description},
	}
}

func (h *Handler) renderCreateBindingError(
	ctx *echo.Context,
	request createRequest,
	err error,
) error {
	if errors.Is(err, binding.ErrUnsupportedMediaType) {
		return h.renderError(
			ctx,
			http.StatusUnsupportedMediaType,
			createErrorResponse(err.Error(), collectionLink()),
		)
	}
	if acceptsHTML(ctx) {
		form := request.form()
		form.SummaryError = "The form could not be read. Try again."
		return h.renderCreateForm(ctx, http.StatusBadRequest, form)
	}
	return h.renderError(
		ctx,
		http.StatusBadRequest,
		createErrorResponse("invalid request body", collectionLink()),
	)
}

func (h *Handler) renderCreateFailure(
	ctx *echo.Context,
	request createRequest,
	err error,
) error {
	if !acceptsHTML(ctx) {
		return h.renderCreateError(ctx, err)
	}
	form := request.form()
	status := http.StatusUnprocessableEntity
	switch {
	case errors.Is(err, models.ErrNameConflict):
		status = http.StatusConflict
		form.Name.Error = "An active template already uses this name."
	case errors.Is(err, validationerrors.ErrInvalidName):
		form.Name.Error = "Enter a name between 1 and 200 characters."
	case errors.Is(err, validationerrors.ErrInvalidDescription):
		form.Description.Error = "Keep the description within 2000 characters."
	default:
		logging.Default().Error("create chore template", "err", err)
		status = http.StatusInternalServerError
		form.SummaryError = "The template could not be created. Try again."
	}
	return h.renderCreateForm(ctx, status, form)
}

func (h *Handler) renderCreateError(ctx *echo.Context, err error) error {
	if errors.Is(err, models.ErrNameConflict) {
		return h.renderError(
			ctx,
			http.StatusConflict,
			createErrorResponse(err.Error(), api.Relations{}),
		)
	}
	if isInvalidCreateRequestError(err) {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			createErrorResponse(err.Error(), collectionLink()),
		)
	}
	logging.Default().Error("create chore template", "err", err)
	return h.renderError(
		ctx,
		http.StatusInternalServerError,
		errorResponse{Error: "something went wrong", Links: collectionLink()},
	)
}

func isInvalidCreateRequestError(err error) bool {
	return errors.Is(err, validationerrors.ErrInvalidName) ||
		errors.Is(err, validationerrors.ErrInvalidDescription)
}

func createErrorResponse(message string, links api.Relations) errorResponse {
	return errorResponse{
		Error:   message,
		Links:   links,
		Actions: api.Relations{createAction()},
	}
}

func createAction() api.Relation {
	return api.Relation{Rel: "create", Href: choreTemplateCollectionHref}
}
