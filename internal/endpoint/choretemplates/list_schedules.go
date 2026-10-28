package choretemplates

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	schedulemodels "github.com/bpsoos/shiftbell/internal/models/schedules"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/choretemplates"
	"github.com/labstack/echo/v5"
)

const inlineScheduleLimit = 10

func (h *Handler) ListSchedules(ctx *echo.Context) error {
	ctx.Response().Header().Set(echo.HeaderVary, "Accept, HX-Request, HX-Target")
	if !acceptsHTML(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	id, err := parseChoreTemplateID(ctx)
	if err != nil {
		return h.renderScheduleListError(ctx, err)
	}
	state := schedulemodels.ScheduleFilter(ctx.QueryParam("state"))
	if state == "" {
		state = schedulemodels.ScheduleFilterActive
	}
	result, err := h.service.ListSchedules(ctx.Request().Context(), id, state)
	if err != nil {
		return h.renderScheduleListError(ctx, err)
	}
	model := scheduleListViewModel(
		id,
		state,
		result.Schedules,
		inlineScheduleRequest(ctx),
	)
	model.TemplateName = result.TemplateName
	return h.renderScheduleList(ctx, http.StatusOK, model)
}

func inlineScheduleRequest(ctx *echo.Context) bool {
	return !fullPage(ctx) &&
		ctx.Request().Header.Get("HX-Target") == "template-schedule-list"
}

func scheduleListViewModel(
	id int,
	state schedulemodels.ScheduleFilter,
	schedules []models.ScheduleReference,
	inline bool,
) viewmodels.ScheduleList {
	model := viewmodels.ScheduleList{
		Heading:      "Active schedules",
		EmptyMessage: "No active schedules.",
		TemplateHref: "/chore-templates/" + strconv.Itoa(id),
		State:        string(state),
		Total:        len(schedules),
		Inline:       inline,
	}
	if state == schedulemodels.ScheduleFilterDeactivated {
		model.Heading = "Deactivated schedules"
		model.EmptyMessage = "No deactivated schedules."
	}
	if inline && len(schedules) > inlineScheduleLimit {
		model.ViewAllHref = model.TemplateHref + "/schedules?state=" + string(state)
		schedules = schedules[:inlineScheduleLimit]
	}
	model.Schedules = make([]viewmodels.ScheduleLink, 0, len(schedules))
	for _, schedule := range schedules {
		model.Schedules = append(model.Schedules, viewmodels.ScheduleLink{
			Name: schedule.Name, Href: "/schedules/" + strconv.Itoa(schedule.Id),
		})
	}
	return model
}

func (h *Handler) renderScheduleListError(ctx *echo.Context, err error) error {
	status := http.StatusInternalServerError
	message := "The schedules could not be loaded. Try again."
	switch {
	case errors.Is(err, models.ErrNotFound):
		status = http.StatusNotFound
		message = "This template could not be found."
	case errors.Is(err, validationerrors.ErrInvalidFilter):
		status = http.StatusUnprocessableEntity
		message = "Invalid schedule state."
	case errors.Is(err, errInvalidChoreTemplateID):
		status = http.StatusUnprocessableEntity
		message = err.Error()
	default:
		logging.Default().Error("list template schedules", "err", err)
	}
	return h.renderScheduleList(ctx, status, viewmodels.ScheduleList{
		Heading: "Unable to load schedules",
		Error:   message,
		Inline:  inlineScheduleRequest(ctx),
	})
}

func (h *Handler) renderScheduleList(
	ctx *echo.Context,
	status int,
	model viewmodels.ScheduleList,
) error {
	return hypermedia.HTML(ctx, status, h.view.ScheduleList(model, fullPage(ctx)))
}
