package shiftbellapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type GetScheduleResult struct {
	StatusCode      int
	SuccessResponse *ScheduleRepresentation
	ErrorResponse   *ErrorResponse
}

func (c *APIClient) GetSchedule(
	ctx context.Context,
	href string,
) (*GetScheduleResult, error) {
	request, err := c.newRequest(ctx, http.MethodGet, href, nil)
	if err != nil {
		return nil, fmt.Errorf("create get schedule request: %w", err)
	}
	statusCode, _, body, err := c.doResponse(request)
	if err != nil {
		return nil, err
	}
	result := &GetScheduleResult{StatusCode: statusCode}
	if statusCode != http.StatusOK {
		result.ErrorResponse, err = decodeErrorResponse(body)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	var schedule ScheduleRepresentation
	if err := json.Unmarshal(body, &schedule); err != nil {
		return nil, fmt.Errorf("decode get schedule response: %w", err)
	}
	result.SuccessResponse = &schedule
	return result, nil
}
