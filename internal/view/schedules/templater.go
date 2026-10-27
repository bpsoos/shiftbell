package schedules

import (
	"github.com/a-h/templ"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/schedules"
	"github.com/bpsoos/shiftbell/internal/view/layouts"
)

type Templater struct{}

func NewTemplater() *Templater {
	return &Templater{}
}

func (t *Templater) Collection(_ viewmodels.Collection, fullPage bool) templ.Component {
	return layouts.Frame("schedules", fullPage, templ.Raw(""))
}

func (t *Templater) Detail(_ viewmodels.Detail, fullPage bool) templ.Component {
	return layouts.Frame("schedules", fullPage, templ.Raw(""))
}

func (t *Templater) Error(_ viewmodels.Error, fullPage bool) templ.Component {
	return layouts.Frame("schedules", fullPage, templ.Raw(""))
}
