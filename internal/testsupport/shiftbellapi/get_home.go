package shiftbellapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *APIClient) GetHome(ctx context.Context) (*GetHomeResult, error) {
	request, err := c.newRequest(ctx, http.MethodGet, "/", nil)
	if err != nil {
		return nil, fmt.Errorf("create get home request: %w", err)
	}
	_, responseBody, err := c.do(request, http.StatusOK)
	if err != nil {
		return nil, err
	}
	var home Home
	if err := json.Unmarshal(responseBody, &home); err != nil {
		return nil, fmt.Errorf("decode home response: %w", err)
	}
	return &GetHomeResult{Home: home}, nil
}
