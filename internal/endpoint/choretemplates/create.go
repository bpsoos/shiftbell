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
		action := createAction()
		return hypermedia.JSON(
			ctx,
			http.StatusBadRequest,
			errorResponse{
				Error:   "invalid request body",
				Links:   collectionLink(),
				Actions: api.Relations{action},
			},
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
			action := createAction()
			return hypermedia.JSON(ctx, http.StatusConflict, errorResponse{
				Error:   err.Error(),
				Links:   api.Relations{},
				Actions: api.Relations{action},
			})
		}
		if isInvalidCreateRequestError(err) {
			action := createAction()
			return hypermedia.JSON(
				ctx,
				http.StatusUnprocessableEntity,
				errorResponse{
					Error:   err.Error(),
					Links:   collectionLink(),
					Actions: api.Relations{action},
				},
			)
		}
		logging.Default().Error("create chore template", "err", err)
		return hypermedia.JSON(
			ctx,
			http.StatusInternalServerError,
			errorResponse{Error: "something went wrong", Links: collectionLink()},
		)
	}

	representation := newRepresentation(choreTemplate)
	ctx.Response().Header().Set(echo.HeaderLocation, representation.Links.Href("self"))
	return hypermedia.JSON(ctx, http.StatusCreated, representation)
}

func isInvalidCreateRequestError(err error) bool {
	return errors.Is(err, validationerrors.ErrInvalidName) ||
		errors.Is(err, validationerrors.ErrInvalidDescription)
}

func createAction() api.Relation {
	return api.Relation{Rel: "create", Href: "/chore-templates"}
}
