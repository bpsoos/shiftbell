package shiftbellapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *APIClient) GetChoreTemplate(
	ctx context.Context,
	params GetChoreTemplateParams,
) (*GetChoreTemplateResult, error) {
	request, err := c.newRequest(ctx, http.MethodGet, params.Link.Href, nil)
	if err != nil {
		return nil, fmt.Errorf("create get chore template request: %w", err)
	}
	statusCode, _, responseBody, err := c.doResponse(request)
	if err != nil {
		return nil, err
	}
	result := &GetChoreTemplateResult{StatusCode: statusCode}
	if statusCode != http.StatusOK {
		result.ErrorResponse, err = decodeErrorResponse(responseBody)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	var representation choreTemplateRepresentation
	if err := json.Unmarshal(responseBody, &representation); err != nil {
		return nil, fmt.Errorf("decode get chore template response: %w", err)
	}
	result.SuccessResponse = &GetChoreTemplateResponse{
		ChoreTemplate: representation.ChoreTemplate,
		Actions:       representation.Actions,
	}
	return result, nil
}
