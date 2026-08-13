package chores

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/bpsoos/shiftbell/internal/endpoint/binding"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	"github.com/labstack/echo/v5"
)

type editChoreRequest struct {
	Name                    string `form:"name"                       json:"name"`
	Description             string `form:"description"                json:"description"`
	Deadline                string `form:"deadline"                   json:"deadline"`
	AlsoUpdateChoreTemplate bool   `form:"also_update_chore_template" json:"also_update_chore_template"`
}

func (h *Handler) editVendorJSON(ctx *echo.Context) error {
	id, err := strconv.Atoi(ctx.ParamOr("id", ""))
	if err != nil || id <= 0 {
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: "invalid chore id"},
		)
	}
	var request editChoreRequest
	if err := json.NewDecoder(ctx.Request().Body).Decode(&request); err != nil {
		return hypermedia.JSON(
			ctx,
			http.StatusBadRequest,
			apiErrorResponse{Error: "invalid JSON"},
		)
	}
	deadline, err := time.Parse(time.DateOnly, request.Deadline)
	if err != nil {
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: validationerrors.ErrInvalidDeadline.Error()},
		)
	}
	existing, err := h.service.Get(ctx.Request().Context(), id)
	if err != nil {
		return h.editChoreError(ctx, err)
	}
	edited, err := h.service.Edit(ctx.Request().Context(), &choremodels.EditChoreParams{
		Id:                      id,
		ScheduleId:              existing.ScheduleId,
		Name:                    request.Name,
		Description:             request.Description,
		Deadline:                deadline,
		AlsoUpdateChoreTemplate: request.AlsoUpdateChoreTemplate,
	})
	if err != nil {
		return h.editChoreError(ctx, err)
	}
	response := newChoreResponse(edited)
	return hypermedia.JSON(ctx, http.StatusOK, choreRepresentation{
		Response: response,
		Actions:  actionsForChore(edited),
	})
}

func (h *Handler) editHTMX(ctx *echo.Context) error {
	id, err := strconv.Atoi(ctx.ParamOr("id", ""))
	if err != nil || id <= 0 {
		return h.renderError(ctx, http.StatusUnprocessableEntity, apiErrorResponse{
			Error: "invalid chore id",
			Links: api.Relations{{Rel: "collection", Href: choreCollectionHref}},
		})
	}

	var request editChoreRequest
	if err := binding.Bind(ctx, &request); err != nil {
		if errors.Is(err, binding.ErrUnsupportedMediaType) {
			return h.renderError(ctx, http.StatusUnsupportedMediaType, apiErrorResponse{
				Error: binding.ErrUnsupportedMediaType.Error(),
			})
		}
		return h.renderError(ctx, http.StatusBadRequest, apiErrorResponse{
			Error: "invalid request body",
		})
	}

	existing, err := h.service.Get(ctx.Request().Context(), id)
	if err != nil {
		return h.renderEditLoadError(ctx, err)
	}
	if existing.Status == choremodels.ChoreStatusCompleted {
		return h.renderNonEditableChoreError(
			ctx,
			existing,
			"completed chore cannot be edited",
		)
	}
	if actionsForChore(existing).Href("edit") == "" {
		return h.renderNonEditableChoreError(ctx, existing, "chore cannot be edited")
	}

	deadline := existing.Deadline
	if existing.ScheduleId == nil {
		deadline, err = time.Parse(time.DateOnly, request.Deadline)
		if err != nil {
			return h.renderEditValidationError(
				ctx,
				existing,
				request,
				validationerrors.ErrInvalidDeadline,
			)
		}
	} else {
		request.Deadline = existing.Deadline.Format(time.DateOnly)
	}

	edited, err := h.service.Edit(ctx.Request().Context(), &choremodels.EditChoreParams{
		Id:                      id,
		ScheduleId:              existing.ScheduleId,
		Name:                    request.Name,
		Description:             request.Description,
		Deadline:                deadline,
		AlsoUpdateChoreTemplate: request.AlsoUpdateChoreTemplate,
	})
	if err != nil {
		if isEditChoreValidationError(err) {
			return h.renderEditValidationError(ctx, existing, request, err)
		}
		if errors.Is(err, choremodels.ErrNotFound) {
			return h.renderEditLoadError(ctx, err)
		}
		logging.Default().Error("edit chore", "err", err)
		return h.renderError(ctx, http.StatusInternalServerError, apiErrorResponse{
			Error: "something went wrong",
			Links: newChoreResponse(existing).Links,
		})
	}
	if edited == nil {
		logging.Default().Error("edit chore returned no chore")
		return h.renderError(ctx, http.StatusInternalServerError, apiErrorResponse{
			Error: "something went wrong",
			Links: newChoreResponse(existing).Links,
		})
	}

	response := newChoreResponse(edited)
	ctx.Response().Header().Set("HX-Replace-Url", response.Links.Href("self"))
	return h.renderDetail(
		ctx,
		http.StatusOK,
		choreRepresentation{
			Response: response,
			Actions:  actionsForChore(edited),
		},
		"Chore updated.",
	)
}

func (h *Handler) renderEditValidationError(
	ctx *echo.Context,
	chore *choremodels.ChoreDetails,
	request editChoreRequest,
	err error,
) error {
	selfHref := newChoreResponse(chore).Links.Href("self")
	fieldErrors := editFieldErrors(err)
	message := err.Error()
	for _, fieldError := range fieldErrors {
		message = fieldError
		break
	}
	feedback := formFeedback{
		Values: map[string]string{
			"name":        request.Name,
			"description": request.Description,
			"deadline":    request.Deadline,
			"also_update_chore_template": strconv.FormatBool(
				request.AlsoUpdateChoreTemplate,
			),
		},
		FieldErrors: fieldErrors,
		Error:       apiErrorResponse{Error: message},
		Action:      api.Relation{Rel: "edit", Href: selfHref},
		CancelHref:  selfHref,
	}
	return h.renderEditForm(
		ctx,
		http.StatusUnprocessableEntity,
		editFormErrorViewModel(chore, feedback),
	)
}

func editFieldErrors(err error) map[string]string {
	switch {
	case errors.Is(err, validationerrors.ErrInvalidName):
		return map[string]string{"name": validationerrors.ErrInvalidName.Error()}
	case errors.Is(err, validationerrors.ErrInvalidDescription):
		return map[string]string{
			"description": validationerrors.ErrInvalidDescription.Error(),
		}
	case errors.Is(err, validationerrors.ErrInvalidDeadline):
		return map[string]string{"deadline": validationerrors.ErrInvalidDeadline.Error()}
	default:
		return nil
	}
}

func (h *Handler) renderNonEditableChoreError(
	ctx *echo.Context,
	chore *choremodels.ChoreDetails,
	message string,
) error {
	return h.renderError(ctx, http.StatusUnprocessableEntity, apiErrorResponse{
		Error: message,
		Links: newChoreResponse(chore).Links,
	})
}

func (h *Handler) editChoreError(
	ctx *echo.Context,
	err error,
) error {
	if isEditChoreValidationError(err) {
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: err.Error()},
		)
	}
	if errors.Is(err, choremodels.ErrNotFound) {
		return hypermedia.JSON(
			ctx,
			http.StatusNotFound,
			apiErrorResponse{Error: choremodels.ErrNotFound.Error()},
		)
	}
	logging.Default().Error("edit chore", "err", err)
	return hypermedia.JSON(
		ctx,
		http.StatusInternalServerError,
		apiErrorResponse{Error: "something went wrong"},
	)
}

func isEditChoreValidationError(err error) bool {
	return errors.Is(err, validationerrors.ErrInvalidName) ||
		errors.Is(err, validationerrors.ErrInvalidDescription) ||
		errors.Is(err, validationerrors.ErrInvalidDeadline)
}
