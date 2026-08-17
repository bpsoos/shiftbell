package chores

import (
	api "github.com/bpsoos/shiftbell/internal/models/api"
	choreapimodels "github.com/bpsoos/shiftbell/internal/models/api/chores"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
)

type CollectionItem struct {
	Chore        choreapimodels.Response
	CompleteHref string
}

type Collection struct {
	Items           []CollectionItem
	Links           api.Relations
	CreateHref      string
	Status          choremodels.ChoreStatus
	Search          string
	SearchOpen      bool
	Notice          string
	AutofocusSearch bool
}

func (model Collection) HasPagination() bool {
	return model.Links.Href("previous") != "" ||
		model.Links.Href("next") != ""
}
