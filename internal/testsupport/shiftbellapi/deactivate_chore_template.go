package shiftbellapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *APIClient) DeactivateChoreTemplate(
	ctx context.Context,
	requestParams RequestParams,
) (*DeactivateChoreTemplateResult, error) {
	request, err := c.newRequest(
		ctx,
		requestParams.Method,
		requestParams.Href,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create deactivate chore template request: %w", err)
	}
	if requestParams.ContentType != "" {
		request.Header.Set("Content-Type", requestParams.ContentType)
	}
	_, responseBody, err := c.do(request, http.StatusOK)
	if err != nil {
		return nil, err
	}
	var representation choreTemplateRepresentation
	if err := json.Unmarshal(responseBody, &representation); err != nil {
		return nil, fmt.Errorf("decode deactivate chore template response: %w", err)
	}
	return &DeactivateChoreTemplateResult{
		ChoreTemplate: representation.ChoreTemplate,
		Actions:       representation.Actions,
	}, nil
}
