package shiftbellapi

import (
	"context"
	"fmt"
	"net/http"
)

func (c *APIClient) ClearChoreAssignment(
	ctx context.Context,
	href string,
) (*GetChoreResult, error) {
	request, err := c.newRequest(ctx, http.MethodDelete, href, nil)
	if err != nil {
		return nil, fmt.Errorf("create clear assignment request: %w", err)
	}
	return c.assignmentResponse(request)
}
