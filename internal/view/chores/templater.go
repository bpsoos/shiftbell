package chores

import (
	"time"

	"github.com/a-h/templ"
	choreviewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
	"github.com/bpsoos/shiftbell/internal/view/layouts"
)

type Config struct {
	AppTimezone *time.Location
	Now         func() time.Time
}

type Templater struct {
	timezone *time.Location
	now      func() time.Time
}

func NewTemplater(config Config) *Templater {
	return &Templater{timezone: config.AppTimezone, now: config.Now}
}

func (t *Templater) Collection(
	model choreviewmodels.Collection,
	fullPage bool,
) templ.Component {
	today := t.today()
	return layouts.Frame("chores", fullPage, collectionContent(t, model, today))
}

func (t *Templater) Detail(
	model choreviewmodels.Detail,
	fullPage bool,
) templ.Component {
	today := t.today()
	return layouts.Frame("chores", fullPage, detailContent(t, model, today))
}

func (t *Templater) Creation(
	model choreviewmodels.Creation,
	fullPage bool,
) templ.Component {
	return layouts.Frame("chores", fullPage, creationContent(model))
}

func (t *Templater) ManualOneOffForm(
	form choreviewmodels.ManualOneOffForm,
	fullPage bool,
) templ.Component {
	if form.NeedsDefaultDeadline() {
		form.Deadline.Value = t.localDate()
	}
	return layouts.Frame("chores", fullPage, manualOneOffFormContent(form))
}

func (t *Templater) TemplateOneOffForm(
	form choreviewmodels.TemplateOneOffForm,
	fullPage bool,
) templ.Component {
	if form.NeedsDefaultDeadline() {
		form.Deadline.Value = t.localDate()
	}
	return layouts.Frame("chores", fullPage, templateOneOffFormContent(form))
}

func (t *Templater) Error(
	model choreviewmodels.Error,
	fullPage bool,
) templ.Component {
	return layouts.Frame("chores", fullPage, errorContent(model))
}

func (t *Templater) localDate() string {
	return t.now().In(t.timezone).Format(time.DateOnly)
}

func (t *Templater) today() time.Time {
	now := t.now().In(t.timezone)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, t.timezone)
}

func (t *Templater) deadlineBadgeVariant(value string, today time.Time) string {
	deadline, err := time.ParseInLocation(time.DateOnly, value, t.timezone)
	if err != nil {
		return "secondary"
	}
	if !deadline.After(today) {
		return "danger"
	}
	if !deadline.After(today.AddDate(0, 0, 7)) {
		return "warning"
	}
	return "secondary"
}

func formControlClass(className string, errorMessage string) string {
	if errorMessage != "" {
		className += " is-invalid"
	}
	return className
}
