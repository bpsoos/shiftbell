package shiftbellapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *APIClient) EditChore(
	ctx context.Context,
	href string,
	params EditChoreParams,
) (*EditChoreResult, error) {
	requestBody, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("encode edit chore request: %w", err)
	}
	request, err := c.newRequest(
		ctx,
		http.MethodPatch,
		href,
		bytes.NewReader(requestBody),
	)
	if err != nil {
		return nil, fmt.Errorf("create edit chore request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	statusCode, _, responseBody, err := c.doResponse(request)
	if err != nil {
		return nil, err
	}
	result := &EditChoreResult{StatusCode: statusCode}
	if statusCode != http.StatusOK {
		result.ErrorResponse, err = decodeErrorResponse(responseBody)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	var representation choreRepresentation
	if err := json.Unmarshal(responseBody, &representation); err != nil {
		return nil, fmt.Errorf("decode edit chore response: %w", err)
	}
	result.SuccessResponse = &EditChoreResponse{
		Chore:   representation.Chore,
		Actions: representation.Actions,
	}
	return result, nil
}
