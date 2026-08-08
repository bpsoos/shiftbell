package choretemplates

import choretemplateapimodels "github.com/bpsoos/shiftbell/internal/models/api/choretemplates"

type Collection struct {
	Collection choretemplateapimodels.CollectionResponse
}

func (model Collection) HasPagination() bool {
	return model.Collection.Links["previous"].Href != "" ||
		model.Collection.Links["next"].Href != ""
}

type Picker struct {
	Collection choretemplateapimodels.PickerCollectionResponse
	BackHref   string
	ManualHref string
}

func (model Picker) HasPagination() bool {
	return model.Collection.Links["previous"].Href != "" ||
		model.Collection.Links["next"].Href != ""
}
