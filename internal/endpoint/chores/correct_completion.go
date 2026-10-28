package chores

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/bpsoos/shiftbell/internal/endpoint/binding"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	"github.com/labstack/echo/v5"
)

func (h *Handler) CorrectCompletion(ctx *echo.Context) error {
	if !supported(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	id, err := parseChoreID(ctx)
	if err != nil {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: err.Error()},
		)
	}
	request, err := parseCorrectionRequest(ctx)
	if err != nil {
		if errors.Is(err, binding.ErrUnsupportedMediaType) {
			return h.renderError(
				ctx,
				http.StatusUnsupportedMediaType,
				apiErrorResponse{Error: err.Error()},
			)
		}
		return h.renderError(
			ctx,
			http.StatusBadRequest,
			apiErrorResponse{Error: "invalid JSON"},
		)
	}
	completedOn, err := time.Parse(time.DateOnly, request.CompletedOn)
	if err != nil {
		return h.renderCorrectionFailure(
			ctx,
			id,
			request.CompletedOn,
			validationerrors.ErrInvalidCompletionDate,
		)
	}
	corrected, err := h.service.CorrectCompletion(
		ctx.Request().Context(),
		&choremodels.CorrectCompletionParams{Id: id, CompletedOn: completedOn},
	)
	if err != nil {
		return h.renderCorrectionFailure(ctx, id, request.CompletedOn, err)
	}
	if corrected == nil {
		logging.Default().Error("correct chore completion returned no chore")
		return h.renderError(
			ctx,
			http.StatusInternalServerError,
			apiErrorResponse{Error: "something went wrong"},
		)
	}
	if acceptsHTML(ctx) {
		ctx.Response().Header().Set("HX-Replace-Url", choreHref(id))
	}
	return h.renderDetail(
		ctx,
		http.StatusOK,
		newChoreRepresentation(corrected),
		"Completion date corrected.",
	)
}

func parseCorrectionRequest(ctx *echo.Context) (completeChoreRequest, error) {
	var request completeChoreRequest
	if ctx.Request().Header.Get(echo.HeaderContentType) == "" {
		err := json.NewDecoder(ctx.Request().Body).Decode(&request)
		return request, err
	}
	err := binding.Bind(ctx, &request)
	return request, err
}

func (h *Handler) renderCorrectionFailure(
	ctx *echo.Context,
	id int,
	completedOn string,
	err error,
) error {
	if !acceptsHTML(ctx) {
		return h.renderCorrectionError(ctx, err)
	}
	if errors.Is(err, choremodels.ErrNotFound) {
		return h.renderCorrectionError(ctx, err)
	}
	form := correctionFormModel(id, completedOn)
	if errors.Is(err, validationerrors.ErrInvalidCompletionDate) {
		form.CompletedOn.Error = "Enter a valid date no later than today."
		return h.renderCompletionCorrectionForm(ctx, http.StatusUnprocessableEntity, form)
	}
	logging.Default().Error("correct chore completion", "err", err)
	form.SummaryError = "The completion date could not be updated. Try again."
	return h.renderCompletionCorrectionForm(ctx, http.StatusInternalServerError, form)
}

func (h *Handler) renderCorrectionError(ctx *echo.Context, err error) error {
	if errors.Is(err, validationerrors.ErrInvalidCompletionDate) {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: validationerrors.ErrInvalidCompletionDate.Error()},
		)
	}
	if errors.Is(err, choremodels.ErrNotFound) {
		return h.renderError(
			ctx,
			http.StatusNotFound,
			apiErrorResponse{Error: choremodels.ErrNotFound.Error()},
		)
	}
	logging.Default().Error("correct chore completion", "err", err)
	return h.renderError(
		ctx,
		http.StatusInternalServerError,
		apiErrorResponse{Error: "something went wrong"},
	)
}
