package schedules

import (
	"strconv"
	"time"

	"github.com/a-h/templ"
	apimodels "github.com/bpsoos/shiftbell/internal/models/api/schedules"
	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	confirmationmodels "github.com/bpsoos/shiftbell/internal/models/view/confirmation"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/schedules"
	confirmationview "github.com/bpsoos/shiftbell/internal/view/confirmation"
	"github.com/bpsoos/shiftbell/internal/view/layouts"
)

type Config struct {
	AppTimezone *time.Location
}

type Templater struct {
	appTimezone *time.Location
}

func NewTemplater(config Config) *Templater {
	return &Templater{appTimezone: config.AppTimezone}
}

func (t *Templater) Collection(
	model viewmodels.Collection,
	fullPage bool,
) templ.Component {
	if fullPage && model.SearchOpen {
		model.AutofocusSearch = true
	}
	return layouts.Frame("schedules", fullPage, t.appTimezone.String(), collection(model))
}

func (t *Templater) Detail(model viewmodels.Detail, fullPage bool) templ.Component {
	return layouts.Frame("schedules", fullPage, t.appTimezone.String(), detail(t, model))
}

func (t *Templater) Error(model viewmodels.Error, fullPage bool) templ.Component {
	return layouts.Frame(
		"schedules",
		fullPage,
		t.appTimezone.String(),
		errorContent(model),
	)
}

func (t *Templater) EditForm(model viewmodels.EditForm, fullPage bool) templ.Component {
	return layouts.Frame("schedules", fullPage, t.appTimezone.String(), editForm(model))
}

func (t *Templater) ConfirmationDialog(model confirmationmodels.Dialog) templ.Component {
	return confirmationview.Dialog(model)
}

func (t *Templater) formatTime(value time.Time) string {
	return value.In(t.appTimezone).Format("2006.01.02. 15:04")
}

func intervalLabel(days int) string {
	if days == 1 {
		return "Every day"
	}
	return "Every " + strconv.Itoa(days) + " days"
}

func scheduleCollectionViewHref(filter models.ScheduleFilter) string {
	if filter == models.ScheduleFilterDeactivated {
		return "/schedules?state=deactivated"
	}
	return "/schedules"
}

func scheduleSearchHref(filter models.ScheduleFilter) string {
	if filter == models.ScheduleFilterDeactivated {
		return "/schedules?state=deactivated&search="
	}
	return "/schedules?search="
}

func scheduleSearchPlaceholder(filter models.ScheduleFilter) string {
	if filter == models.ScheduleFilterDeactivated {
		return "Search deactivated schedules"
	}
	return "Search schedules"
}

func detailCollectionHref(schedule apimodels.Representation) string {
	href := schedule.Links.Href("collection")
	if href == "" || schedule.DeactivatedAt == nil {
		return href
	}
	return href + "?state=deactivated"
}
