package shiftbellapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *APIClient) DeactivateChoreTemplate(
	ctx context.Context,
	params DeactivateChoreTemplateParams,
) (*DeactivateChoreTemplateResult, error) {
	request, err := c.newRequest(ctx, params.Action.Method, params.Action.Href, nil)
	if err != nil {
		return nil, fmt.Errorf("create deactivate chore template request: %w", err)
	}
	if params.Action.ContentType != "" {
		request.Header.Set("Content-Type", params.Action.ContentType)
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
