package chores

import choreapimodels "github.com/bpsoos/shiftbell/internal/models/api/chores"

type Collection struct {
	Collection choreapimodels.CollectionResponse
}

func (model Collection) HasPagination() bool {
	return model.Collection.Links["previous"].Href != "" ||
		model.Collection.Links["next"].Href != ""
}
