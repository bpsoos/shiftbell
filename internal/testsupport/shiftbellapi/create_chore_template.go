package shiftbellapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *APIClient) CreateChoreTemplate(
	ctx context.Context,
	href string,
	params CreateChoreTemplateParams,
) (*CreateChoreTemplateResult, error) {
	requestBody, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("encode create chore template request: %w", err)
	}
	request, err := c.newRequest(
		ctx,
		http.MethodPost,
		href,
		bytes.NewReader(requestBody),
	)
	if err != nil {
		return nil, fmt.Errorf("create chore template request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	statusCode, header, responseBody, err := c.doResponse(request)
	if err != nil {
		return nil, err
	}
	result := &CreateChoreTemplateResult{StatusCode: statusCode}
	if statusCode != http.StatusCreated {
		result.ErrorResponse, err = decodeErrorResponse(responseBody)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	var representation choreTemplateRepresentation
	if err := json.Unmarshal(responseBody, &representation); err != nil {
		return nil, fmt.Errorf("decode create chore template response: %w", err)
	}
	result.SuccessResponse = &CreateChoreTemplateResponse{
		ChoreTemplate: representation.ChoreTemplate,
		Actions:       representation.Actions,
		Location:      header.Get("Location"),
	}
	return result, nil
}
