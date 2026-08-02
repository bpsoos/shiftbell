package shiftbellapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *APIClient) GetChoreCreationStep(
	ctx context.Context,
	href string,
) (*GetChoreCreationStepResult, error) {
	request, err := c.newRequest(ctx, http.MethodGet, href, nil)
	if err != nil {
		return nil, fmt.Errorf("create get chore creation step request: %w", err)
	}
	_, responseBody, err := c.do(request, http.StatusOK)
	if err != nil {
		return nil, err
	}
	var step ChoreCreationStep
	if err := json.Unmarshal(responseBody, &step); err != nil {
		return nil, fmt.Errorf("decode chore creation step response: %w", err)
	}
	return &GetChoreCreationStepResult{Step: step}, nil
}
