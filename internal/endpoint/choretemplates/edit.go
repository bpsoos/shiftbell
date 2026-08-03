package choretemplates

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	"github.com/labstack/echo/v5"
)

type editRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *Handler) Edit(ctx *echo.Context) error {
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
			errorResponse{Error: "invalid JSON"},
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
			selfHref := fmt.Sprintf("/chore-templates/%d", id)
			return hypermedia.JSON(ctx, http.StatusConflict, errorResponse{
				Error: err.Error(),
				Links: map[string]hypermedia.Link{},
				Actions: map[string]hypermedia.Action{
					"edit": activeActions(selfHref)["edit"],
				},
			})
		}
		if errors.Is(err, validationerrors.ErrInvalidName) ||
			errors.Is(err, validationerrors.ErrInvalidDescription) {
			return hypermedia.JSON(
				ctx,
				http.StatusUnprocessableEntity,
				errorResponse{Error: err.Error()},
			)
		}
		if errors.Is(err, models.ErrNotFound) {
			return hypermedia.JSON(
				ctx,
				http.StatusNotFound,
				errorResponse{Error: models.ErrNotFound.Error()},
			)
		}
		logging.Default().Error("edit chore template", "err", err)
		return hypermedia.JSON(
			ctx,
			http.StatusInternalServerError,
			errorResponse{Error: "something went wrong"},
		)
	}
	return hypermedia.JSON(ctx, http.StatusOK, newRepresentation(edited))
}
