package chores

import choreapimodels "github.com/bpsoos/shiftbell/internal/models/api/chores"

type Collection struct {
	Collection choreapimodels.CollectionResponse
}

func (model Collection) HasPagination() bool {
	return model.Collection.Links.Href("previous") != "" ||
		model.Collection.Links.Href("next") != ""
}
