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
	statusCode, _, responseBody, err := c.doResponse(request)
	if err != nil {
		return nil, err
	}
	result := &GetChoreCreationStepResult{StatusCode: statusCode}
	if statusCode != http.StatusOK {
		result.ErrorResponse, err = decodeErrorResponse(responseBody)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	var step ChoreCreationStep
	if err := json.Unmarshal(responseBody, &step); err != nil {
		return nil, fmt.Errorf("decode chore creation step response: %w", err)
	}
	result.SuccessResponse = &step
	return result, nil
}
