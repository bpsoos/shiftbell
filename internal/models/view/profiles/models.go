package profiles

import apimodels "github.com/bpsoos/shiftbell/internal/models/api/profiles"

type Field struct {
	Value string
	Error string
}

type Form struct {
	ActionHref   string
	CancelHref   string
	Name         Field
	SummaryError string
	Setup        bool
}

type Collection struct {
	Profiles   []apimodels.Response
	CreateHref string
}

type Error struct {
	Message        string
	CollectionHref string
}
