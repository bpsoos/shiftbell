package choretemplates

import choretemplateapimodels "github.com/bpsoos/shiftbell/internal/models/api/choretemplates"

type Collection struct {
	Collection choretemplateapimodels.CollectionResponse
	Notice     string
}

func (model Collection) HasPagination() bool {
	return model.Collection.Links.Href("previous") != "" ||
		model.Collection.Links.Href("next") != ""
}

type Picker struct {
	Collection choretemplateapimodels.PickerCollectionResponse
	BackHref   string
	ManualHref string
}

func (model Picker) HasPagination() bool {
	return model.Collection.Links.Href("previous") != "" ||
		model.Collection.Links.Href("next") != ""
}
