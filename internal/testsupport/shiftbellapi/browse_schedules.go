package shiftbellapi

import (
	"context"
	"fmt"
	"net/url"
)

type BrowseSchedulesParams struct {
	Href   string
	Search string
	State  string
}

func (c *APIClient) BrowseSchedules(
	ctx context.Context,
	params BrowseSchedulesParams,
) (*ScheduleCollection, error) {
	reference, err := url.Parse(params.Href)
	if err != nil {
		return nil, fmt.Errorf("parse schedule collection reference: %w", err)
	}
	query := reference.Query()
	if params.Search != "" {
		query.Set("search", params.Search)
	}
	if params.State != "" {
		query.Set("state", params.State)
	}
	reference.RawQuery = query.Encode()
	return c.GetSchedules(ctx, reference.String())
}
