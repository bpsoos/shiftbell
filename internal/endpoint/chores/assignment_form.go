package chores

import (
	"errors"
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	models "github.com/bpsoos/shiftbell/internal/models/chores"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
	"github.com/labstack/echo/v5"
)

func (h *Handler) AssignmentForm(ctx *echo.Context) error {
	if !h.profilesUIEnabled {
		return ctx.NoContent(http.StatusNotFound)
	}
	if !acceptsHTML(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	ctx.Response().Header().Set("Cache-Control", "private, no-store")
	id, err := parseChoreID(ctx)
	if err != nil {
		return h.renderAssignmentForm(
			ctx,
			http.StatusUnprocessableEntity,
			viewmodels.Assignment{Error: err.Error()},
		)
	}
	chore, err := h.service.Get(ctx.Request().Context(), id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return h.renderAssignmentForm(
				ctx,
				http.StatusNotFound,
				viewmodels.Assignment{Error: "Chore not found."},
			)
		}
		logging.Default().Error("load chore assignment", "err", err)
		return h.renderAssignmentForm(
			ctx,
			http.StatusInternalServerError,
			viewmodels.Assignment{
				Error:     "Could not load assignment. Try again.",
				RetryHref: choreHref(id) + "/assignment",
			},
		)
	}
	profiles, err := h.profileService.List(ctx.Request().Context())
	if err != nil {
		logging.Default().Error("list assignment profiles", "err", err)
		return h.renderAssignmentForm(
			ctx,
			http.StatusInternalServerError,
			viewmodels.Assignment{
				Error:     "Could not load profiles. Try again.",
				RetryHref: actionsForChore(chore).Href("set_assignment"),
			},
		)
	}
	return h.renderAssignmentForm(ctx, http.StatusOK, viewmodels.Assignment{
		Chore: newChoreRepresentation(chore), Profiles: profiles,
	})
}

func (h *Handler) renderAssignmentForm(
	ctx *echo.Context,
	status int,
	model viewmodels.Assignment,
) error {
	return hypermedia.HTML(ctx, status, h.view.Assignment(model))
}
