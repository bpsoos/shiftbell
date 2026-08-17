package choretemplates

import (
	choretemplateapimodels "github.com/bpsoos/shiftbell/internal/models/api/choretemplates"
	choretemplatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
)

type Collection struct {
	Collection      choretemplateapimodels.CollectionResponse
	Filter          choretemplatemodels.ChoreTemplateFilter
	Search          string
	SearchOpen      bool
	Notice          string
	AutofocusSearch bool
}

func (model Collection) HasPagination() bool {
	return model.Collection.Links.Href("previous") != "" ||
		model.Collection.Links.Href("next") != ""
}

type Picker struct {
	Collection      choretemplateapimodels.PickerCollectionResponse
	BackHref        string
	ManualHref      string
	Search          string
	AutofocusSearch bool
}

func (model Picker) HasPagination() bool {
	return model.Collection.Links.Href("previous") != "" ||
		model.Collection.Links.Href("next") != ""
}
