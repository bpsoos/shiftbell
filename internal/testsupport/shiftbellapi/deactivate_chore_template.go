package shiftbellapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *APIClient) DeactivateChoreTemplate(
	ctx context.Context,
	href string,
) (*DeactivateChoreTemplateResult, error) {
	request, err := c.newRequest(
		ctx,
		http.MethodPut,
		href,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create deactivate chore template request: %w", err)
	}
	statusCode, _, responseBody, err := c.doResponse(request)
	if err != nil {
		return nil, err
	}
	result := &DeactivateChoreTemplateResult{StatusCode: statusCode}
	if statusCode != http.StatusOK {
		result.ErrorResponse, err = decodeErrorResponse(responseBody)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	var representation choreTemplateRepresentation
	if err := json.Unmarshal(responseBody, &representation); err != nil {
		return nil, fmt.Errorf("decode deactivate chore template response: %w", err)
	}
	result.ChoreTemplate = representation.ChoreTemplate
	result.Actions = representation.Actions
	return result, nil
}
