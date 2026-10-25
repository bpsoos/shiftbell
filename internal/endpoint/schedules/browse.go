package schedules

import (
	"errors"
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
			map[string]string{"error": err.Error()},
		)
	}

	page, err := h.service.Browse(ctx.Request().Context(), &params)
	if err != nil {
		return h.renderBrowseError(ctx, err)
	}

	return hypermedia.JSON(
		ctx,
		http.StatusOK,
		browseResponse(page, browseLinks(ctx.Request().URL, params, page.More)),
	)
}

func parseBrowseRequest(ctx *echo.Context) (models.BrowseSchedulesParams, error) {
	params := models.BrowseSchedulesParams{Limit: 20}
	var state string
	err := echo.QueryParamsBinder(ctx).
		String("state", &state).
		Int("offset", &params.Offset).
		BindError()
	if err != nil || params.Offset < 0 {
		return models.BrowseSchedulesParams{}, validationerrors.ErrInvalidOffset
	}
	params.Filter = models.ScheduleFilter(state)
	return params, nil
}

func (h *Handler) renderBrowseError(ctx *echo.Context, err error) error {
	if errors.Is(err, validationerrors.ErrInvalidFilter) {
		return hypermedia.JSON(ctx, http.StatusUnprocessableEntity,
			map[string]string{"error": validationerrors.ErrInvalidFilter.Error()})
	}
	logging.Default().Error("browse schedules", "err", err)
	return hypermedia.JSON(ctx, http.StatusInternalServerError,
		map[string]string{"error": "something went wrong"})
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
	links := api.Relations{
		{Rel: "self", Href: requestURL.RequestURI()},
		{Rel: "active", Href: scheduleStateHref(requestURL, models.ScheduleFilterActive)},
		{
			Rel:  "deactivated",
			Href: scheduleStateHref(requestURL, models.ScheduleFilterDeactivated),
		},
	}
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

func scheduleStateHref(requestURL *url.URL, state models.ScheduleFilter) string {
	stateURL := *requestURL
	query := stateURL.Query()
	query.Set("state", string(state))
	query.Del("offset")
	stateURL.RawQuery = query.Encode()
	return stateURL.RequestURI()
}
