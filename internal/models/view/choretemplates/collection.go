package choretemplates

import (
	choretemplatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
)

type Collection struct {
	Items           []CollectionItem
	PreviousHref    string
	NextHref        string
	Filter          choretemplatemodels.ChoreTemplateFilter
	Search          string
	SearchOpen      bool
	Notice          string
	AutofocusSearch bool
}

func (model Collection) HasPagination() bool {
	return model.PreviousHref != "" || model.NextHref != ""
}

type CollectionItem struct {
	Name        string
	Description string
	DetailHref  string
}

type Picker struct {
	Items           []PickerItem
	PreviousHref    string
	NextHref        string
	BackHref        string
	ManualHref      string
	Search          string
	AutofocusSearch bool
}

type PickerItem struct {
	Name       string
	SelectHref string
}

func (model Picker) HasPagination() bool {
	return model.PreviousHref != "" || model.NextHref != ""
}
