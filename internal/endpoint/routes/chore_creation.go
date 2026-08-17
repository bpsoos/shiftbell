package routes

import (
	"net/url"
	"strconv"
)

type ChoreCreationSource string

const ChoreCreationSourceManual ChoreCreationSource = "manual"

type ChoreCreationRecurrence string

const (
	ChoreCreationRecurrenceOneOff    ChoreCreationRecurrence = "one-off"
	ChoreCreationRecurrenceScheduled ChoreCreationRecurrence = "scheduled"
)

type ChoreCreation struct {
	Source          ChoreCreationSource
	ChoreTemplateId int
	Recurrence      ChoreCreationRecurrence
}

func (route ChoreCreation) Href() string {
	target := url.URL{Path: "/chores/new"}
	query := target.Query()
	if route.Source != "" {
		query.Set("source", string(route.Source))
	}
	if route.ChoreTemplateId != 0 {
		query.Set("template_id", strconv.Itoa(route.ChoreTemplateId))
	}
	if route.Recurrence != "" {
		query.Set("recurrence", string(route.Recurrence))
	}
	target.RawQuery = query.Encode()
	return target.RequestURI()
}
