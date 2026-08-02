package shiftbellapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *APIClient) GetChoreTemplates(ctx context.Context, href string) (*GetChoreTemplatesResult, error) {
	request, err := c.newRequest(ctx, http.MethodGet, href, nil)
	if err != nil {
		return nil, fmt.Errorf("create get chore templates request: %w", err)
	}
	_, responseBody, err := c.do(request, http.StatusOK)
	if err != nil {
		return nil, err
	}
	var collection ChoreTemplateCollection
	if err := json.Unmarshal(responseBody, &collection); err != nil {
		return nil, fmt.Errorf("decode chore template collection response: %w", err)
	}
	return &GetChoreTemplatesResult{Collection: collection}, nil
}
