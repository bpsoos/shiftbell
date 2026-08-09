package shiftbellapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *APIClient) CorrectChoreCompletion(
	ctx context.Context,
	href string,
	params CorrectChoreCompletionParams,
) (*CorrectChoreCompletionResult, error) {
	requestBody, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("encode correct chore completion request: %w", err)
	}
	request, err := c.newRequest(
		ctx,
		http.MethodPatch,
		href,
		bytes.NewReader(requestBody),
	)
	if err != nil {
		return nil, fmt.Errorf("create correct chore completion request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	statusCode, _, responseBody, err := c.doResponse(request)
	if err != nil {
		return nil, err
	}
	result := &CorrectChoreCompletionResult{StatusCode: statusCode}
	if statusCode != http.StatusOK {
		result.ErrorResponse, err = decodeErrorResponse(responseBody)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	var representation choreRepresentation
	if err := json.Unmarshal(responseBody, &representation); err != nil {
		return nil, fmt.Errorf("decode correct chore completion response: %w", err)
	}
	result.SuccessResponse = &CorrectChoreCompletionResponse{
		Chore:   representation.Chore,
		Actions: representation.Actions,
	}
	return result, nil
}
