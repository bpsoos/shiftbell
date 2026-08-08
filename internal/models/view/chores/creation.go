package chores

import choreapimodels "github.com/bpsoos/shiftbell/internal/models/api/chores"

type Creation struct {
	Creation           choreapimodels.CreationResponse
	BackHref           string
	SpecifyNewHref     string
	SelectTemplateHref string
	OneOffHref         string
	ScheduledHref      string
}

func (model Creation) HasSourceOptions() bool {
	return model.SpecifyNewHref != "" || model.SelectTemplateHref != ""
}

func (model Creation) HasRecurrenceOptions() bool {
	return model.OneOffHref != "" || model.ScheduledHref != ""
}
