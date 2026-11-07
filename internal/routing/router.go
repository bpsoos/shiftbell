package routing

import "github.com/labstack/echo/v5"

type HomeHandler interface {
	Get(*echo.Context) error
}

type ProfileHandler interface {
	SelectionStatus(*echo.Context) error
	Selection(*echo.Context) error
	Select(*echo.Context) error
	List(*echo.Context) error
	Create(*echo.Context) error
	CreateForm(*echo.Context) error
	Setup(*echo.Context) error
	SetupStatus(*echo.Context) error
}

type ScheduleHandler interface {
	Browse(*echo.Context) error
	Get(*echo.Context) error
	Edit(*echo.Context) error
	Deactivate(*echo.Context) error
	EditForm(*echo.Context) error
	ConfirmDeactivation(*echo.Context) error
}

type ChoreTemplateHandler interface {
	ListSchedules(*echo.Context) error
	CreateForm(*echo.Context) error
	Browse(*echo.Context) error
	Create(*echo.Context) error
	Get(*echo.Context) error
	EditForm(*echo.Context) error
	Edit(*echo.Context) error
	ConfirmDeactivation(*echo.Context) error
	Deactivate(*echo.Context) error
}

type ChoreHandler interface {
	SetAssignment(*echo.Context) error
	AssignmentForm(*echo.Context) error
	ClearAssignment(*echo.Context) error
	Get(*echo.Context) error
	GetBatch(*echo.Context) error
	Edit(*echo.Context) error
	Patch(*echo.Context) error
	New(*echo.Context) error
	Create(*echo.Context) error
	ConfirmCompletion(*echo.Context) error
	Complete(*echo.Context) error
	CorrectCompletion(*echo.Context) error
	CompletionCorrectionForm(*echo.Context) error
	ConfirmDeletion(*echo.Context) error
	Delete(*echo.Context) error
}

type RouterDeps struct {
	ProfileHandler       ProfileHandler
	ScheduleHandler      ScheduleHandler
	HomeHandler          HomeHandler
	ChoreTemplateHandler ChoreTemplateHandler
	ChoreHandler         ChoreHandler
}

type Router struct {
	profileHandler       ProfileHandler
	scheduleHandler      ScheduleHandler
	homeHandler          HomeHandler
	choreTemplateHandler ChoreTemplateHandler
	choreHandler         ChoreHandler
}

func NewRouter(deps *RouterDeps) *Router {
	return &Router{
		profileHandler:       deps.ProfileHandler,
		scheduleHandler:      deps.ScheduleHandler,
		homeHandler:          deps.HomeHandler,
		choreTemplateHandler: deps.ChoreTemplateHandler,
		choreHandler:         deps.ChoreHandler,
	}
}
