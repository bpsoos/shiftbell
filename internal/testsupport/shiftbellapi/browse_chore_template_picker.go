package shiftbellapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

func (c *APIClient) BrowseChoreTemplatePicker(
	ctx context.Context,
	params BrowseChoreTemplatePickerParams,
) (*BrowseChoreTemplatePickerResult, error) {
	reference, err := url.Parse(params.Href)
	if err != nil {
		return nil, fmt.Errorf("parse chore template picker reference: %w", err)
	}
	query := reference.Query()
	if params.Search != "" {
		query.Set("search", params.Search)
	}
	if params.Limit > 0 {
		query.Set("limit", strconv.Itoa(params.Limit))
	}
	reference.RawQuery = query.Encode()
	request, err := c.newRequest(ctx, http.MethodGet, reference.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create browse chore template picker request: %w", err)
	}
	_, responseBody, err := c.do(request, http.StatusOK)
	if err != nil {
		return nil, err
	}
	var collection ChoreTemplatePickerCollection
	if err := json.Unmarshal(responseBody, &collection); err != nil {
		return nil, fmt.Errorf("decode chore template picker response: %w", err)
	}
	return &BrowseChoreTemplatePickerResult{Collection: collection}, nil
}
