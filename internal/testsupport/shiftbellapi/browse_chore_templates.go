package shiftbellapi

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

func (c *APIClient) BrowseChoreTemplates(
	ctx context.Context,
	params BrowseChoreTemplatesParams,
) (*GetChoreTemplatesResult, error) {
	reference, err := url.Parse(params.Href)
	if err != nil {
		return nil, fmt.Errorf("parse chore template collection reference: %w", err)
	}
	query := reference.Query()
	if params.Search != "" {
		query.Set("search", params.Search)
	}
	if params.State != "" {
		query.Set("state", params.State)
	}
	if params.Limit > 0 {
		query.Set("limit", strconv.Itoa(params.Limit))
	}
	reference.RawQuery = query.Encode()
	return c.GetChoreTemplates(ctx, reference.String())
}
