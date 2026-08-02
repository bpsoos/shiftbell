package chores

import (
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	"github.com/labstack/echo/v5"
)

func (h *Handler) browse(ctx *echo.Context) error {
	page, err := h.service.Browse(
		ctx.Request().Context(),
		&choremodels.BrowseChoresParams{
			Status: choremodels.ChoreStatusActive,
			Offset: 0,
			Limit:  20,
		},
	)
	if err != nil {
		logging.Default().Error("browse chores", "err", err)
		return hypermedia.JSON(
			ctx,
			http.StatusInternalServerError,
			apiErrorResponse{Error: "something went wrong"},
		)
	}

	items := make([]choreResponse, len(page.Chores))
	for i := range page.Chores {
		items[i] = newChoreResponse(&page.Chores[i])
	}

	return hypermedia.JSON(ctx, http.StatusOK, choreCollectionResponse{
		Items: items,
		More:  page.More,
		Links: map[string]hypermedia.Link{
			"self": {Href: ctx.Request().URL.RequestURI()},
		},
		Actions: map[string]hypermedia.Action{
			"create": createChoreNavigationAction(),
		},
	})
}
