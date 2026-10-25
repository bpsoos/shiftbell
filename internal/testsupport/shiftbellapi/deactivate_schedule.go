package shiftbellapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type DeactivateScheduleResult struct {
	StatusCode      int
	SuccessResponse *ScheduleRepresentation
	ErrorResponse   *ErrorResponse
}

func (c *APIClient) DeactivateSchedule(
	ctx context.Context,
	href string,
) (*DeactivateScheduleResult, error) {
	request, err := c.newRequest(ctx, http.MethodPut, href, nil)
	if err != nil {
		return nil, fmt.Errorf("create deactivate schedule request: %w", err)
	}
	statusCode, _, body, err := c.doResponse(request)
	if err != nil {
		return nil, err
	}
	result := &DeactivateScheduleResult{StatusCode: statusCode}
	if statusCode != http.StatusOK {
		result.ErrorResponse, err = decodeErrorResponse(body)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	var schedule ScheduleRepresentation
	if err := json.Unmarshal(body, &schedule); err != nil {
		return nil, fmt.Errorf("decode deactivate schedule response: %w", err)
	}
	result.SuccessResponse = &schedule
	return result, nil
}
