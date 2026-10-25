package schedules

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	"github.com/labstack/echo/v5"
)

type collectionResponse struct {
	Items []response    `json:"items"`
	More  bool          `json:"more"`
	Links api.Relations `json:"_links"`
}

func (h *Handler) Browse(ctx *echo.Context) error {
	if hypermedia.Negotiate(ctx.Request()) != hypermedia.RepresentationJSON {
		return hypermedia.NotAcceptable(ctx)
	}

	params, err := parseBrowseRequest(ctx)
	if err != nil {
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			map[string]string{"error": "invalid offset"},
		)
	}

	page, err := h.service.Browse(ctx.Request().Context(), &params)
	if err != nil {
		logging.Default().Error("browse schedules", "err", err)
		return hypermedia.JSON(
			ctx,
			http.StatusInternalServerError,
			map[string]string{"error": "something went wrong"},
		)
	}

	return hypermedia.JSON(
		ctx,
		http.StatusOK,
		browseResponse(page, browseLinks(ctx.Request().URL, params, page.More)),
	)
}

func parseBrowseRequest(ctx *echo.Context) (models.BrowseSchedulesParams, error) {
	params := models.BrowseSchedulesParams{Limit: 20}
	err := echo.QueryParamsBinder(ctx).
		Int("offset", &params.Offset).
		BindError()
	if err != nil || params.Offset < 0 {
		return models.BrowseSchedulesParams{}, validationerrors.ErrInvalidOffset
	}
	return params, nil
}

func browseResponse(page *models.SchedulePage, links api.Relations) collectionResponse {
	items := make([]response, len(page.Schedules))
	for i, schedule := range page.Schedules {
		items[i] = newResponse(&schedule)
	}
	return collectionResponse{Items: items, More: page.More, Links: links}
}

func browseLinks(
	requestURL *url.URL,
	params models.BrowseSchedulesParams,
	more bool,
) api.Relations {
	links := api.Relations{{Rel: "self", Href: requestURL.RequestURI()}}
	if more {
		links = append(
			links,
			api.Relation{
				Rel:  "next",
				Href: schedulePageHref(requestURL, params.Offset+params.Limit),
			},
		)
	}
	if params.Offset > 0 {
		links = append(
			links,
			api.Relation{
				Rel:  "previous",
				Href: schedulePageHref(requestURL, max(0, params.Offset-params.Limit)),
			},
		)
	}
	return links
}

func schedulePageHref(requestURL *url.URL, offset int) string {
	pageURL := *requestURL
	query := pageURL.Query()
	query.Set("offset", strconv.Itoa(offset))
	pageURL.RawQuery = query.Encode()
	return pageURL.RequestURI()
}
