package chores

import (
	api "github.com/bpsoos/shiftbell/internal/models/api"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
)

type Response struct {
	Id          int                     `json:"id"`
	ScheduleId  *int                    `json:"schedule_id"`
	Status      choremodels.ChoreStatus `json:"status"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Deadline    string                  `json:"deadline"`
	CompletedOn *string                 `json:"completed_on"`
	Links       map[string]api.Link     `json:"_links"`
}

type Representation struct {
	Response
	Actions map[string]api.Action `json:"_actions"`
}

type CollectionResponse struct {
	Items   []Response            `json:"items"`
	More    bool                  `json:"more"`
	Links   map[string]api.Link   `json:"_links"`
	Actions map[string]api.Action `json:"_actions"`
}

type CreationChoice struct {
	Label string `json:"label"`
	Href  string `json:"href"`
}

type CreationTemplate struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type CreationResponse struct {
	Step     string            `json:"step"`
	Template *CreationTemplate `json:"template,omitempty"`
	Choices  []CreationChoice  `json:"choices,omitempty"`
	Fields   []api.ActionField `json:"fields,omitempty"`
	Action   *api.Action       `json:"action,omitempty"`
}
