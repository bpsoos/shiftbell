package chores

import (
	"errors"
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	confirmationviewmodels "github.com/bpsoos/shiftbell/internal/models/view/confirmation"
	"github.com/labstack/echo/v5"
)

func (h *Handler) confirmDeletion(ctx *echo.Context) error {
	id, err := parseChoreID(ctx)
	if err != nil {
		return hypermedia.NoContent(ctx, http.StatusUnprocessableEntity)
	}
	chore, err := h.service.Get(ctx.Request().Context(), id)
	if err != nil {
		return renderDeletionLoadError(ctx, err)
	}
	if actionsForChore(chore).Href("delete") == "" {
		return hypermedia.NoContent(ctx, http.StatusUnprocessableEntity)
	}
	return h.renderConfirmationDialog(
		ctx,
		http.StatusOK,
		deletionDialogViewModel(chore, ""),
	)
}

func deletionDialogViewModel(
	chore *choremodels.ChoreDetails,
	errorMessage string,
) confirmationviewmodels.Dialog {
	return confirmationviewmodels.Dialog{
		Heading:      "Delete chore?",
		Prompt:       "Delete",
		Name:         chore.Name,
		Supporting:   []string{"This cannot be undone."},
		ActionHref:   newChoreResponse(chore).Links.Href("self"),
		ActionMethod: "delete",
		ActionLabel:  "Delete permanently",
		Error:        errorMessage,
		Icon:         "trash",
	}
}

func renderDeletionLoadError(ctx *echo.Context, err error) error {
	if errors.Is(err, choremodels.ErrNotFound) {
		return hypermedia.NoContent(ctx, http.StatusNotFound)
	}
	logging.Default().Error("load chore for deletion", "err", err)
	return hypermedia.NoContent(ctx, http.StatusInternalServerError)
}
