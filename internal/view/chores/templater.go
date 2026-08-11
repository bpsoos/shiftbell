package chores

import (
	"fmt"
	"strings"
	"time"

	"github.com/a-h/templ"
	choreviewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
	"github.com/bpsoos/shiftbell/internal/view/layouts"
)

type deadlineGroup struct {
	Label  string
	State  string
	Chores []choreviewmodels.CollectionItem
}

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

func (t *Templater) CompletionDialog(
	model choreviewmodels.CompletionDialog,
) templ.Component {
	today := t.localDate()
	model.CompletedOnMax = today
	if model.CompletedOn.Value == "" && model.CompletedOn.Error == "" {
		model.CompletedOn.Value = today
	}
	return completionDialog(model)
}

func (t *Templater) localDate() string {
	return t.now().In(t.timezone).Format(time.DateOnly)
}

func (t *Templater) today() time.Time {
	now := t.now().In(t.timezone)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, t.timezone)
}

func (t *Templater) deadlineGroups(
	chores []choreviewmodels.CollectionItem,
	today time.Time,
) []deadlineGroup {
	groups := []deadlineGroup{
		{Label: "Overdue", State: "overdue"},
		{Label: "Today", State: "today"},
		{Label: "Upcoming", State: "upcoming"},
	}
	for _, chore := range chores {
		switch t.deadlineState(chore.Chore.Deadline, today) {
		case "overdue":
			groups[0].Chores = append(groups[0].Chores, chore)
		case "today":
			groups[1].Chores = append(groups[1].Chores, chore)
		default:
			groups[2].Chores = append(groups[2].Chores, chore)
		}
	}
	nonempty := make([]deadlineGroup, 0, len(groups))
	for _, group := range groups {
		if len(group.Chores) != 0 {
			nonempty = append(nonempty, group)
		}
	}
	return nonempty
}

func (t *Templater) deadlineState(value string, today time.Time) string {
	deadline, err := time.ParseInLocation(time.DateOnly, value, t.timezone)
	if err != nil {
		return "upcoming"
	}
	if deadline.Before(today) {
		return "overdue"
	}
	if deadline.Equal(today) {
		return "today"
	}
	return "upcoming"
}

func (t *Templater) relativeDeadline(value string, today time.Time) string {
	deadline, err := time.ParseInLocation(time.DateOnly, value, t.timezone)
	if err != nil {
		return value
	}
	days := calendarDayDifference(deadline, today)
	switch {
	case days < 0:
		return fmt.Sprintf("%dd late", -days)
	case days == 0:
		return "Today"
	case days < 14:
		return fmt.Sprintf("%dd", days)
	case days < 28:
		return fmt.Sprintf("%dw", days/7)
	case deadline.Year() == today.Year():
		return deadline.Format("2 Jan")
	default:
		return deadline.Format("2 Jan 2006")
	}
}

func (t *Templater) exactDeadline(value string) string {
	deadline, err := time.ParseInLocation(time.DateOnly, value, t.timezone)
	if err != nil {
		return value
	}
	return deadline.Format("Monday, 2 January 2006")
}

func (t *Templater) deadlineAccessibleLabel(value string) string {
	deadline, err := time.ParseInLocation(time.DateOnly, value, t.timezone)
	if err != nil {
		return "Due " + value
	}
	return "Due " + deadline.Format("2 January 2006")
}

func calendarDayDifference(deadline time.Time, today time.Time) int {
	deadlineUTC := time.Date(
		deadline.Year(), deadline.Month(), deadline.Day(), 0, 0, 0, 0, time.UTC,
	)
	todayUTC := time.Date(
		today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC,
	)
	return int(deadlineUTC.Sub(todayUTC) / (24 * time.Hour))
}

func oneOffCreationHref(href string) string {
	if strings.Contains(href, "?") {
		return href + "&recurrence=one-off"
	}
	return href + "?recurrence=one-off"
}

func formControlClass(className string, errorMessage string) string {
	if errorMessage != "" {
		className += " is-invalid"
	}
	return className
}
