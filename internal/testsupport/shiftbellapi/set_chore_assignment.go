package shiftbellapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *APIClient) SetChoreAssignment(
	ctx context.Context,
	href string,
	owner int,
) (*GetChoreResult, error) {
	body, err := json.Marshal(struct {
		OwnerProfileId int `json:"owner_profile_id"`
	}{OwnerProfileId: owner})
	if err != nil {
		return nil, fmt.Errorf("encode assignment request: %w", err)
	}
	request, err := c.newRequest(ctx, http.MethodPut, href, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create assignment request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	return c.assignmentResponse(request)
}

func (c *APIClient) assignmentResponse(request *http.Request) (*GetChoreResult, error) {
	status, _, response, err := c.doResponse(request)
	if err != nil {
		return nil, err
	}
	result := &GetChoreResult{StatusCode: status}
	if status != http.StatusOK {
		result.ErrorResponse, err = decodeErrorResponse(response)
		return result, err
	}
	var representation choreRepresentation
	if err := json.Unmarshal(response, &representation); err != nil {
		return nil, fmt.Errorf("decode assignment response: %w", err)
	}
	result.SuccessResponse = &GetChoreResponse{
		Chore:   representation.Chore,
		Actions: representation.Actions,
	}
	return result, nil
}
