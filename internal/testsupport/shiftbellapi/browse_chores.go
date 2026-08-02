package shiftbellapi

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

func (c *APIClient) BrowseChores(
	ctx context.Context,
	params BrowseChoresParams,
) (*GetChoresResult, error) {
	reference, err := url.Parse(params.Href)
	if err != nil {
		return nil, fmt.Errorf("parse chore collection reference: %w", err)
	}
	query := reference.Query()
	if params.Search != "" {
		query.Set("search", params.Search)
	}
	if params.Status != "" {
		query.Set("status", params.Status)
	}
	if params.Limit > 0 {
		query.Set("limit", strconv.Itoa(params.Limit))
	}
	reference.RawQuery = query.Encode()
	return c.GetChores(ctx, reference.String())
}
