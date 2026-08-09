package shiftbellapi

import (
	"context"
	"fmt"
	"net/http"
)

func (c *APIClient) DeleteChore(
	ctx context.Context,
	href string,
) (*DeleteChoreResult, error) {
	request, err := c.newRequest(
		ctx,
		http.MethodDelete,
		href,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create delete chore request: %w", err)
	}
	statusCode, _, responseBody, err := c.doResponse(request)
	if err != nil {
		return nil, err
	}
	result := &DeleteChoreResult{StatusCode: statusCode}
	if statusCode != http.StatusNoContent {
		result.ErrorResponse, err = decodeErrorResponse(responseBody)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
