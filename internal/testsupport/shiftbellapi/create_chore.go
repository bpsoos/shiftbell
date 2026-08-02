package shiftbellapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *APIClient) CreateChore(
	ctx context.Context,
	requestParams RequestParams,
	params CreateChoreParams,
) (*CreateChoreResult, error) {
	requestBody, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("encode create chore request: %w", err)
	}
	request, err := c.newRequest(
		ctx,
		requestParams.Method,
		requestParams.Href,
		bytes.NewReader(requestBody),
	)
	if err != nil {
		return nil, fmt.Errorf("create chore request: %w", err)
	}
	request.Header.Set("Content-Type", requestParams.ContentType)
	statusCode, header, responseBody, err := c.doResponse(request)
	if err != nil {
		return nil, err
	}
	result := &CreateChoreResult{StatusCode: statusCode}
	if statusCode != http.StatusCreated {
		result.ErrorResponse, err = decodeErrorResponse(responseBody)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	var representation choreRepresentation
	if err := json.Unmarshal(responseBody, &representation); err != nil {
		return nil, fmt.Errorf("decode create chore response: %w", err)
	}
	result.SuccessResponse = &CreateChoreResponse{
		Chore:    representation.Chore,
		Actions:  representation.Actions,
		Location: header.Get("Location"),
	}
	return result, nil
}
