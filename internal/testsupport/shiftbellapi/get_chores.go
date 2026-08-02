package shiftbellapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *APIClient) GetChores(
	ctx context.Context,
	href string,
) (*GetChoresResult, error) {
	request, err := c.newRequest(ctx, http.MethodGet, href, nil)
	if err != nil {
		return nil, fmt.Errorf("create get chores request: %w", err)
	}
	_, responseBody, err := c.do(request, http.StatusOK)
	if err != nil {
		return nil, err
	}
	var collection ChoreCollection
	if err := json.Unmarshal(responseBody, &collection); err != nil {
		return nil, fmt.Errorf("decode chore collection response: %w", err)
	}
	return &GetChoresResult{Collection: collection}, nil
}
