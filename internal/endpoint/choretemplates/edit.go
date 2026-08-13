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
	"github.com/labstack/echo/v5"
)

type editRequest struct {
	Name        string `form:"name"        json:"name"`
	Description string `form:"description" json:"description"`
}

func (h *Handler) Edit(ctx *echo.Context) error {
	if hypermedia.Accepts(ctx.Request()) {
		return h.editVendorJSON(ctx)
	}
	ctx.Response().Header().Set(echo.HeaderVary, "Accept, HX-Request")
	if !acceptsHTMXHTML(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	return h.editHTMX(ctx)
}

func (h *Handler) editVendorJSON(ctx *echo.Context) error {
	id, err := parseChoreTemplateID(ctx)
	if err != nil {
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			errorResponse{Error: err.Error()},
		)
	}
	var request editRequest
	if err := json.NewDecoder(ctx.Request().Body).Decode(&request); err != nil {
		return hypermedia.JSON(
			ctx,
			http.StatusBadRequest,
			editErrorResponse(id, "invalid request body"),
		)
	}
	edited, err := h.service.Edit(
		ctx.Request().Context(),
		request.params(id),
	)
	if err != nil {
		return h.renderVendorJSONEditError(ctx, id, err)
	}
	return hypermedia.JSON(ctx, http.StatusOK, newRepresentation(edited))
}

func (h *Handler) editHTMX(ctx *echo.Context) error {
	id, err := parseChoreTemplateID(ctx)
	if err != nil {
		return h.renderError(ctx, http.StatusUnprocessableEntity, errorResponse{
			Error: err.Error(),
			Links: collectionLink(),
		})
	}

	var request editRequest
	if err := binding.Bind(ctx, &request); err != nil {
		return h.renderEditBindingError(ctx, id, err)
	}

	edited, err := h.service.Edit(
		ctx.Request().Context(),
		request.params(id),
	)
	if err != nil {
		return h.renderHTMXEditError(ctx, id, request, err)
	}
	if edited == nil {
		return h.renderMissingEditResult(ctx, id)
	}
	return h.renderEditedTemplate(ctx, edited)
}

func (request editRequest) params(id int) *models.EditChoreTemplateParams {
	return &models.EditChoreTemplateParams{
		Id:          id,
		Name:        request.Name,
		Description: request.Description,
	}
}

func (h *Handler) renderVendorJSONEditError(
	ctx *echo.Context,
	id int,
	err error,
) error {
	switch {
	case errors.Is(err, models.ErrNameConflict):
		return hypermedia.JSON(
			ctx,
			http.StatusConflict,
			editErrorResponse(id, err.Error()),
		)
	case errors.Is(err, validationerrors.ErrInvalidName),
		errors.Is(err, validationerrors.ErrInvalidDescription):
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			editErrorResponse(id, err.Error()),
		)
	case errors.Is(err, models.ErrInactive):
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			resourceErrorResponse(id, models.ErrInactive.Error()),
		)
	case errors.Is(err, models.ErrNotFound):
		return hypermedia.JSON(
			ctx,
			http.StatusNotFound,
			errorResponse{Error: models.ErrNotFound.Error(), Links: collectionLink()},
		)
	default:
		logging.Default().Error("edit chore template", "err", err)
		return hypermedia.JSON(
			ctx,
			http.StatusInternalServerError,
			resourceErrorResponse(id, "something went wrong"),
		)
	}
}

func (h *Handler) renderEditBindingError(
	ctx *echo.Context,
	id int,
	err error,
) error {
	status := http.StatusBadRequest
	message := "invalid request body"
	if errors.Is(err, binding.ErrUnsupportedMediaType) {
		status = http.StatusUnsupportedMediaType
		message = binding.ErrUnsupportedMediaType.Error()
	}
	return h.renderError(ctx, status, errorResponse{
		Error: message,
		Links: resourceErrorResponse(id, "").Links,
	})
}

func (h *Handler) renderHTMXEditError(
	ctx *echo.Context,
	id int,
	request editRequest,
	err error,
) error {
	switch {
	case errors.Is(err, models.ErrNameConflict):
		return h.renderEditFeedback(
			ctx,
			http.StatusConflict,
			id,
			request,
			"name",
			models.ErrNameConflict.Error(),
		)
	case errors.Is(err, validationerrors.ErrInvalidName):
		return h.renderEditFeedback(
			ctx,
			http.StatusUnprocessableEntity,
			id,
			request,
			"name",
			validationerrors.ErrInvalidName.Error(),
		)
	case errors.Is(err, validationerrors.ErrInvalidDescription):
		return h.renderEditFeedback(
			ctx,
			http.StatusUnprocessableEntity,
			id,
			request,
			"description",
			validationerrors.ErrInvalidDescription.Error(),
		)
	case errors.Is(err, models.ErrInactive):
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			resourceErrorResponse(id, models.ErrInactive.Error()),
		)
	case errors.Is(err, models.ErrNotFound):
		return h.renderError(ctx, http.StatusNotFound, errorResponse{
			Error: models.ErrNotFound.Error(),
			Links: collectionLink(),
		})
	default:
		logging.Default().Error("edit chore template", "err", err)
		return h.renderError(
			ctx,
			http.StatusInternalServerError,
			resourceErrorResponse(id, "something went wrong"),
		)
	}
}

func (h *Handler) renderMissingEditResult(ctx *echo.Context, id int) error {
	logging.Default().Error("edit chore template returned no template")
	return h.renderError(
		ctx,
		http.StatusInternalServerError,
		resourceErrorResponse(id, "something went wrong"),
	)
}

func (h *Handler) renderEditedTemplate(
	ctx *echo.Context,
	edited *models.ChoreTemplate,
) error {
	representation := newRepresentation(edited)
	ctx.Response().Header().Set("HX-Replace-Url", representation.Links.Href("self"))
	return h.renderDetail(
		ctx,
		http.StatusOK,
		representation,
		"Template updated.",
	)
}

func (h *Handler) renderEditFeedback(
	ctx *echo.Context,
	status int,
	id int,
	request editRequest,
	field string,
	message string,
) error {
	return h.renderEditForm(
		ctx,
		status,
		editFormFeedbackViewModel(
			id,
			request,
			map[string]string{field: message},
			message,
		),
	)
}

func editErrorResponse(id int, message string) errorResponse {
	selfHref := resourceHref(id)
	return errorResponse{
		Error: message,
		Links: api.Relations{},
		Actions: api.Relations{
			{Rel: "edit", Href: selfHref},
		},
	}
}

func resourceErrorResponse(id int, message string) errorResponse {
	return errorResponse{
		Error: message,
		Links: api.Relations{
			{Rel: "self", Href: resourceHref(id)},
			{Rel: "collection", Href: choreTemplateCollectionHref},
		},
	}
}
