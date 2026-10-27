package schedules

import (
	"strconv"

	"github.com/a-h/templ"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/schedules"
	"github.com/bpsoos/shiftbell/internal/view/layouts"
)

type Templater struct{}

func NewTemplater() *Templater {
	return &Templater{}
}

func (t *Templater) Collection(
	model viewmodels.Collection,
	fullPage bool,
) templ.Component {
	return layouts.Frame("schedules", fullPage, collection(model))
}

func (t *Templater) Detail(model viewmodels.Detail, fullPage bool) templ.Component {
	return layouts.Frame("schedules", fullPage, detail(model))
}

func (t *Templater) Error(_ viewmodels.Error, fullPage bool) templ.Component {
	return layouts.Frame("schedules", fullPage, templ.Raw(""))
}

func intervalLabel(days int) string {
	if days == 1 {
		return "Every day"
	}
	return "Every " + strconv.Itoa(days) + " days"
}
