package shiftbellapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *APIClient) GetChore(
	ctx context.Context,
	href string,
) (*GetChoreResult, error) {
	request, err := c.newRequest(ctx, http.MethodGet, href, nil)
	if err != nil {
		return nil, fmt.Errorf("create get chore request: %w", err)
	}
	statusCode, _, responseBody, err := c.doResponse(request)
	if err != nil {
		return nil, err
	}
	result := &GetChoreResult{StatusCode: statusCode}
	if statusCode != http.StatusOK {
		result.ErrorResponse, err = decodeErrorResponse(responseBody)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	var representation choreRepresentation
	if err := json.Unmarshal(responseBody, &representation); err != nil {
		return nil, fmt.Errorf("decode get chore response: %w", err)
	}
	result.SuccessResponse = &GetChoreResponse{
		Chore:   representation.Chore,
		Actions: representation.Actions,
	}
	return result, nil
}
