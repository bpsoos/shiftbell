package choretemplates

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

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
	id, err := strconv.Atoi(ctx.ParamOr("id", ""))
	if err != nil || id <= 0 {
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			errorResponse{Error: "invalid chore template id"},
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
		&models.EditChoreTemplateParams{
			Id:          id,
			Name:        request.Name,
			Description: request.Description,
		},
	)
	if err != nil {
		if errors.Is(err, models.ErrNameConflict) {
			return hypermedia.JSON(
				ctx,
				http.StatusConflict,
				editErrorResponse(id, err.Error()),
			)
		}
		if errors.Is(err, validationerrors.ErrInvalidName) ||
			errors.Is(err, validationerrors.ErrInvalidDescription) {
			return hypermedia.JSON(
				ctx,
				http.StatusUnprocessableEntity,
				editErrorResponse(id, err.Error()),
			)
		}
		if errors.Is(err, models.ErrInactive) {
			return hypermedia.JSON(
				ctx,
				http.StatusUnprocessableEntity,
				resourceErrorResponse(id, models.ErrInactive.Error()),
			)
		}
		if errors.Is(err, models.ErrNotFound) {
			return hypermedia.JSON(
				ctx,
				http.StatusNotFound,
				errorResponse{Error: models.ErrNotFound.Error(), Links: collectionLink()},
			)
		}
		logging.Default().Error("edit chore template", "err", err)
		return hypermedia.JSON(
			ctx,
			http.StatusInternalServerError,
			resourceErrorResponse(id, "something went wrong"),
		)
	}
	return hypermedia.JSON(ctx, http.StatusOK, newRepresentation(edited))
}

func (h *Handler) editHTMX(ctx *echo.Context) error {
	id, err := strconv.Atoi(ctx.ParamOr("id", ""))
	if err != nil || id <= 0 {
		return h.renderError(ctx, http.StatusUnprocessableEntity, errorResponse{
			Error: "invalid chore template id",
			Links: collectionLink(),
		})
	}

	var request editRequest
	if err := binding.Bind(ctx, &request); err != nil {
		if errors.Is(err, binding.ErrUnsupportedMediaType) {
			return h.renderError(ctx, http.StatusUnsupportedMediaType, errorResponse{
				Error: binding.ErrUnsupportedMediaType.Error(),
				Links: resourceErrorResponse(id, "").Links,
			})
		}
		return h.renderError(ctx, http.StatusBadRequest, errorResponse{
			Error: "invalid request body",
			Links: resourceErrorResponse(id, "").Links,
		})
	}

	edited, err := h.service.Edit(
		ctx.Request().Context(),
		&models.EditChoreTemplateParams{
			Id:          id,
			Name:        request.Name,
			Description: request.Description,
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNameConflict):
			return h.renderEditFeedback(
				ctx,
				http.StatusConflict,
				id,
				request,
				map[string]string{"name": models.ErrNameConflict.Error()},
				models.ErrNameConflict.Error(),
			)
		case errors.Is(err, validationerrors.ErrInvalidName):
			return h.renderEditFeedback(
				ctx,
				http.StatusUnprocessableEntity,
				id,
				request,
				map[string]string{"name": validationerrors.ErrInvalidName.Error()},
				validationerrors.ErrInvalidName.Error(),
			)
		case errors.Is(err, validationerrors.ErrInvalidDescription):
			return h.renderEditFeedback(
				ctx,
				http.StatusUnprocessableEntity,
				id,
				request,
				map[string]string{
					"description": validationerrors.ErrInvalidDescription.Error(),
				},
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
	if edited == nil {
		logging.Default().Error("edit chore template returned no template")
		return h.renderError(
			ctx,
			http.StatusInternalServerError,
			resourceErrorResponse(id, "something went wrong"),
		)
	}

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
	fieldErrors map[string]string,
	summaryError string,
) error {
	return h.renderEditForm(
		ctx,
		status,
		editFormFeedbackViewModel(id, request, fieldErrors, summaryError),
	)
}

func editErrorResponse(id int, message string) errorResponse {
	selfHref := fmt.Sprintf("/chore-templates/%d", id)
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
			{Rel: "collection", Href: "/chore-templates"},
		},
	}
}

func resourceHref(id int) string {
	return fmt.Sprintf("/chore-templates/%d", id)
}
