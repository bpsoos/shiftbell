package chores

import (
	api "github.com/bpsoos/shiftbell/internal/models/api"
	choreapimodels "github.com/bpsoos/shiftbell/internal/models/api/chores"
)

type CollectionItem struct {
	Chore        choreapimodels.Response
	CompleteHref string
}

type Collection struct {
	Items   []CollectionItem
	Links   api.Relations
	Actions api.Relations
	Notice  string
}

func (model Collection) HasPagination() bool {
	return model.Links.Href("previous") != "" ||
		model.Links.Href("next") != ""
}
