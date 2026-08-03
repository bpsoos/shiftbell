package choretemplates

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
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
		return ctx.NoContent(http.StatusNotAcceptable)
	}

	var request createRequest
	if err := json.NewDecoder(ctx.Request().Body).Decode(&request); err != nil {
		return hypermedia.JSON(
			ctx,
			http.StatusBadRequest,
			errorResponse{Error: "invalid JSON"},
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
		if errors.Is(err, models.ErrNameConflict) {
			return hypermedia.JSON(
				ctx,
				http.StatusConflict,
				errorResponse{
					Error:   err.Error(),
					Links:   map[string]hypermedia.Link{},
					Actions: map[string]hypermedia.Action{"create": createAction()},
				},
			)
		}
		if isInvalidCreateRequestError(err) {
			return hypermedia.JSON(
				ctx,
				http.StatusUnprocessableEntity,
				errorResponse{
					Error:   err.Error(),
					Links:   collectionLink(),
					Actions: map[string]hypermedia.Action{"create": createAction()},
				},
			)
		}
		logging.Default().Error("create chore template", "err", err)
		return hypermedia.JSON(
			ctx,
			http.StatusInternalServerError,
			errorResponse{Error: "something went wrong"},
		)
	}

	representation := newRepresentation(choreTemplate)
	ctx.Response().Header().Set(
		echo.HeaderLocation,
		representation.Links["self"].Href,
	)
	return hypermedia.JSON(ctx, http.StatusCreated, representation)
}

func isInvalidCreateRequestError(err error) bool {
	return errors.Is(err, validationerrors.ErrInvalidName) ||
		errors.Is(err, validationerrors.ErrInvalidDescription)
}

func createAction() hypermedia.Action {
	return hypermedia.Action{
		Href:        "/chore-templates",
		Method:      http.MethodPost,
		ContentType: "application/json",
		Fields: []hypermedia.ActionField{
			{Name: "name", Type: "string", Required: true},
			{Name: "description", Type: "string", Required: false},
		},
	}
}
