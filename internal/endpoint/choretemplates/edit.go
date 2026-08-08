package choretemplates

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	"github.com/labstack/echo/v5"
)

type editRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *Handler) Edit(ctx *echo.Context) error {
	if !hypermedia.Accepts(ctx.Request()) {
		return hypermedia.NotAcceptable(ctx)
	}

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

func editErrorResponse(id int, message string) errorResponse {
	selfHref := fmt.Sprintf("/chore-templates/%d", id)
	return errorResponse{
		Error: message,
		Links: map[string]api.Link{},
		Actions: map[string]api.Action{
			"edit": activeActions(selfHref)["edit"],
		},
	}
}

func resourceErrorResponse(id int, message string) errorResponse {
	return errorResponse{
		Error: message,
		Links: map[string]api.Link{
			"self":       {Href: resourceHref(id)},
			"collection": {Href: "/chore-templates"},
		},
	}
}

func resourceHref(id int) string {
	return fmt.Sprintf("/chore-templates/%d", id)
}
