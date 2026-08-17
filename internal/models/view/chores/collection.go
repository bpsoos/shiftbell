package chores

import choremodels "github.com/bpsoos/shiftbell/internal/models/chores"

type CollectionItem struct {
	Status       choremodels.ChoreStatus
	Name         string
	Description  string
	Deadline     string
	CompletedOn  *string
	DetailHref   string
	CompleteHref string
}

type Collection struct {
	Items           []CollectionItem
	SelfHref        string
	PreviousHref    string
	NextHref        string
	CreateHref      string
	Status          choremodels.ChoreStatus
	Search          string
	SearchOpen      bool
	Notice          string
	AutofocusSearch bool
}

func (model Collection) HasPagination() bool {
	return model.PreviousHref != "" || model.NextHref != ""
}
