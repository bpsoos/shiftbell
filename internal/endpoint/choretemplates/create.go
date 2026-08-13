package choretemplates

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	"github.com/labstack/echo/v5"
)

type createRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *Handler) Create(ctx *echo.Context) error {
	if !hypermedia.Accepts(ctx.Request()) {
		return hypermedia.NotAcceptable(ctx)
	}

	var request createRequest
	if err := json.NewDecoder(ctx.Request().Body).Decode(&request); err != nil {
		return hypermedia.JSON(
			ctx,
			http.StatusBadRequest,
			createErrorResponse("invalid request body", collectionLink()),
		)
	}

	choreTemplate, err := h.service.Create(
		ctx.Request().Context(),
		&models.CreateChoreTemplateParams{
			Name:        request.Name,
			Description: request.Description,
		},
	)
	if err != nil {
		return h.renderCreateError(ctx, err)
	}

	representation := newRepresentation(choreTemplate)
	ctx.Response().Header().Set(echo.HeaderLocation, representation.Links.Href("self"))
	return hypermedia.JSON(ctx, http.StatusCreated, representation)
}

func (h *Handler) renderCreateError(ctx *echo.Context, err error) error {
	if errors.Is(err, models.ErrNameConflict) {
		return hypermedia.JSON(
			ctx,
			http.StatusConflict,
			createErrorResponse(err.Error(), api.Relations{}),
		)
	}
	if isInvalidCreateRequestError(err) {
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			createErrorResponse(err.Error(), collectionLink()),
		)
	}
	logging.Default().Error("create chore template", "err", err)
	return hypermedia.JSON(
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
