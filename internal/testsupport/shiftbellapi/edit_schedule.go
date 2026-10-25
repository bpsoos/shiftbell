package shiftbellapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type EditScheduleParams struct {
	Name         string `json:"name"`
	IntervalDays int    `json:"interval_days"`
}

type EditScheduleResult struct {
	StatusCode      int
	SuccessResponse *ScheduleRepresentation
	ErrorResponse   *ErrorResponse
}

func (c *APIClient) EditSchedule(
	ctx context.Context,
	href string,
	params EditScheduleParams,
) (*EditScheduleResult, error) {
	body, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("encode edit schedule request: %w", err)
	}
	request, err := c.newRequest(ctx, http.MethodPatch, href, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create edit schedule request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	statusCode, _, responseBody, err := c.doResponse(request)
	if err != nil {
		return nil, err
	}
	result := &EditScheduleResult{StatusCode: statusCode}
	if statusCode != http.StatusOK {
		result.ErrorResponse, err = decodeErrorResponse(responseBody)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	var schedule ScheduleRepresentation
	if err := json.Unmarshal(responseBody, &schedule); err != nil {
		return nil, fmt.Errorf("decode edit schedule response: %w", err)
	}
	result.SuccessResponse = &schedule
	return result, nil
}
