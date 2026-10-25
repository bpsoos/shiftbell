package shiftbellapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type ScheduleCollection struct {
	Items []Schedule `json:"items"`
	More  bool       `json:"more"`
	Links Relations  `json:"_links"`
}

type Schedule struct {
	Id                  int       `json:"id"`
	Name                string    `json:"name"`
	ChoreTemplateId     int       `json:"chore_template_id"`
	ChoreTemplateName   string    `json:"chore_template_name"`
	IntervalDays        int       `json:"interval_days"`
	DeactivatedAt       *string   `json:"deactivated_at"`
	ActiveChoreId       *int      `json:"active_chore_id"`
	ActiveChoreDeadline *string   `json:"active_chore_deadline"`
	Links               Relations `json:"_links"`
}

func (c *APIClient) GetSchedules(
	ctx context.Context,
	href string,
) (*ScheduleCollection, error) {
	request, err := c.newRequest(ctx, http.MethodGet, href, nil)
	if err != nil {
		return nil, fmt.Errorf("create get schedules request: %w", err)
	}
	_, body, err := c.do(request, http.StatusOK)
	if err != nil {
		return nil, err
	}
	var collection ScheduleCollection
	if err := json.Unmarshal(body, &collection); err != nil {
		return nil, fmt.Errorf("decode schedule collection response: %w", err)
	}
	return &collection, nil
}
