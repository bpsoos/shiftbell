package chores

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	choretemplatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	"github.com/labstack/echo/v5"
)

type Templater interface {
	Table(
		context.Context,
		io.Writer,
		int,
		int,
		*choremodels.GetChoreBatchResult,
	) error
	Chore(
		context.Context,
		io.Writer,
		*choremodels.Chore,
	) error
	Page(
		context.Context,
		io.Writer,
		int,
		int,
		*choremodels.GetChoreBatchResult,
		*choretemplatemodels.GetChoreTemplateBatchResult,
		*choretemplatemodels.ChoreTemplate,
	) error
	PageWithLayout(
		context.Context,
		io.Writer,
		int,
		int,
		*choremodels.GetChoreBatchResult,
	) error
	NewChorePage(
		context.Context,
		io.Writer,
	) error
	JoinedComponents(
		ctx context.Context,
		w io.Writer,
		componentSpecifiers ...choremodels.NewChoreTemplateComponent,
	) error
}

type ChoreTemplatePersister interface {
	Get(context.Context, int) (*choretemplatemodels.ChoreTemplateDetails, error)
}

type Persister interface {
	Create(params *choremodels.CreateOneOffChoreParams) (*choremodels.Chore, error)
	GetBatch(offset int, limit int) (*choremodels.GetChoreBatchResult, error)
	Get(context.Context, int) (*choremodels.Chore, error)
	MarkComplete(id int, completedOn time.Time) error
}

type Service interface {
	Browse(
		context.Context,
		*choremodels.BrowseChoresParams,
	) (*choremodels.ChorePage, error)
	Create(
		context.Context,
		*choremodels.CreateChoreParams,
	) (*choremodels.CreateChoreResult, error)
	Complete(
		context.Context,
		*choremodels.CompleteChoreParams,
	) (*choremodels.CompleteChoreResult, error)
	CorrectCompletion(
		context.Context,
		*choremodels.CorrectCompletionParams,
	) (*choremodels.ChoreDetails, error)
	Edit(
		context.Context,
		*choremodels.EditChoreParams,
	) (*choremodels.ChoreDetails, error)
	Delete(context.Context, int) error
	Get(context.Context, int) (*choremodels.ChoreDetails, error)
}

type HandlerDeps struct {
	Templater              Templater
	Persister              Persister
	ChoreTemplatePersister ChoreTemplatePersister
	Service                Service
}

type Handler struct {
	templater              Templater
	persister              Persister
	choreTemplatePersister ChoreTemplatePersister
	service                Service
}

func NewHandler(deps *HandlerDeps) *Handler {
	return &Handler{
		templater:              deps.Templater,
		persister:              deps.Persister,
		choreTemplatePersister: deps.ChoreTemplatePersister,
		service:                deps.Service,
	}
}

func (h *Handler) Get(ctx *echo.Context) error {
	if hypermedia.Accepts(ctx.Request()) {
		return h.get(ctx)
	}

	idStr := ctx.ParamOr("id", "")
	if idStr == "" {
		return ctx.String(http.StatusUnprocessableEntity, "id missing")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logging.Default().Info("invalid id received", "err", err)
		return ctx.String(http.StatusUnprocessableEntity, "invalid id")
	}

	chore, err := h.persister.Get(ctx.Request().Context(), id)
	if err != nil {
		logging.Default().Error("get batch error", "err", err)
		return ctx.String(http.StatusInternalServerError, "something went wrong")
	}

	getForType := ctx.QueryParamOr("for", "readonly")
	switch getForType {
	case "readonly":
		ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		ctx.Response().WriteHeader(http.StatusOK)

		return h.templater.Chore(ctx.Request().Context(), ctx.Response(), chore)
	default:
		logging.Default().Error("unknown get for cause", "get_for_type", getForType)
		return ctx.String(http.StatusUnprocessableEntity, "unknown get for type")
	}
}

func (h *Handler) GetBatch(ctx *echo.Context) error {
	if hypermedia.Accepts(ctx.Request()) {
		return h.browse(ctx)
	}

	offsetStr := ctx.QueryParamOr("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		logging.Default().Info("invalid offset", "err", err)
		return ctx.String(http.StatusUnprocessableEntity, "invalid offset")
	}
	limitStr := ctx.QueryParamOr("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		logging.Default().Info("invalid limit", "err", err)
		return ctx.String(http.StatusUnprocessableEntity, "invalid limit")
	}
	statusFilter := ctx.QueryParamOr("status", "active")
	if statusFilter != "active" {
		logging.Default().Info("unsupported status filter", "status_filter", statusFilter)
		return ctx.String(http.StatusUnprocessableEntity, "unsupported status filter")
	}
	content := ctx.QueryParamOr("content", "all")

	chores, err := h.persister.GetBatch(offset, limit)
	if err != nil {
		logging.Default().Error("get batch error", "err", err)
		return ctx.String(http.StatusInternalServerError, "something went wrong")
	}

	switch content {
	case "table":
		ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		ctx.Response().WriteHeader(http.StatusOK)

		return h.templater.Table(
			ctx.Request().Context(),
			ctx.Response(),
			offset,
			limit,
			chores,
		)
	case "all":
		ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		ctx.Response().WriteHeader(http.StatusOK)

		if ctx.Request().Header.Get("HX-Request") == "true" {
			return h.templater.Page(
				ctx.Request().Context(),
				ctx.Response(),
				offset,
				limit,
				chores,
				nil,
				nil,
			)
		}

		return h.templater.PageWithLayout(
			ctx.Request().Context(),
			ctx.Response(),
			offset,
			limit,
			chores,
		)
	default:
		logging.Default().Error("unknown content", "content", content)
		return ctx.String(http.StatusUnprocessableEntity, "unknown content")
	}
}

func (h *Handler) Patch(ctx *echo.Context) error {
	if hypermedia.Accepts(ctx.Request()) {
		return h.edit(ctx)
	}

	idStr := ctx.ParamOr("id", "")
	if idStr == "" {
		return ctx.String(http.StatusUnprocessableEntity, "id missing")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logging.Default().Info("invalid id received", "err", err)
		return ctx.String(http.StatusUnprocessableEntity, "invalid id")
	}
	status := ctx.FormValueOr("status", "")
	switch status {
	case "completed":
		err = h.persister.MarkComplete(id, time.Now())
		if err != nil {
			logging.Default().Error("patch chore status error", "err", err)
			return ctx.String(http.StatusInternalServerError, "something went wrong")
		}

		ctx.Response().Header().Set("HX-Trigger", "load-chores")
		return ctx.String(http.StatusOK, "OK")
	case "active":
		return ctx.String(http.StatusConflict, "setting status to active dissalowed")
	case "":
		logging.Default().Info("patch request request content empty")
		return ctx.String(http.StatusUnprocessableEntity, "patch content missing")
	default:
		logging.Default().Info("unknown status", "status", status)
		return ctx.String(http.StatusUnprocessableEntity, "unknown status")
	}
}

func (h *Handler) Create(ctx *echo.Context) error {
	if hypermedia.Accepts(ctx.Request()) {
		return h.create(ctx)
	}

	name := ctx.FormValueOr("name", "")
	if name == "" {
		logging.Default().Info("missing name for create chore")
		return ctx.String(http.StatusUnprocessableEntity, "missing name")
	}
	description := ctx.FormValueOr("description", "")
	deadlineValue := ctx.FormValueOr("deadline", "")
	if deadlineValue == "" {
		logging.Default().Info("missing deadline for create chore")
		return ctx.String(http.StatusUnprocessableEntity, "missing deadline")
	}
	deadline, err := time.Parse(time.DateOnly, deadlineValue)
	if err != nil {
		logging.Default().
			Info("invalid deadline for create chore", "deadline", deadlineValue, "err", err)
		return ctx.String(http.StatusUnprocessableEntity, "invalid deadline")
	}

	_, err = h.persister.Create(&choremodels.CreateOneOffChoreParams{
		Name:        name,
		Description: description,
		Deadline:    deadline,
	})
	if err != nil {
		logging.Default().Error("create chore error", "err", err)
		return ctx.String(http.StatusInternalServerError, "something went wrong")
	}

	ctx.Response().Header().Set("HX-Redirect", "/chores")
	return ctx.String(http.StatusOK, "OK")
}

type NewChoreParams struct {
	Fields          []string `query:"field"`
	ChoreTemplateId int      `query:"choreTemplateId"`
	InputType       string   `query:"inputType"`
}

func (h *Handler) New(ctx *echo.Context) error {
	if hypermedia.Accepts(ctx.Request()) {
		return h.newChore(ctx)
	}

	ctxValues := &choremodels.NewChoreCtxValues{
		SelectedChoreTemplate: nil,
		IsManual:              false,
	}
	params := &NewChoreParams{
		Fields:          nil,
		ChoreTemplateId: 0,
		InputType:       "selectChoreTemplate",
	}
	if err := ctx.Bind(params); err != nil {
		return ctx.String(http.StatusUnprocessableEntity, "invalid query param(s)")
	}

	if ctx.QueryParams().Has("inputType") {
		switch params.InputType {
		case "selectChoreTemplate":
			ctxValues.IsManual = false
		case "manual":
			ctxValues.IsManual = true
		default:
			logging.Default().Info("invalid input type", "input_type", params.InputType)
			return ctx.String(http.StatusUnprocessableEntity, "invalid input type")
		}
	}

	if ctx.QueryParams().Has("choreTemplateId") {
		selectedChoreTemplate, err := h.choreTemplatePersister.Get(
			ctx.Request().Context(),
			params.ChoreTemplateId,
		)
		if err != nil {
			logging.Default().Info("get chore template error", "err", err)
			return ctx.String(http.StatusInternalServerError, "something went wrong")
		}
		logging.Default().
			Info("fetched chore template successfully", "chore_template_id", params.ChoreTemplateId)
		ctxValues.SelectedChoreTemplate = &selectedChoreTemplate.ChoreTemplate
	}

	if ctx.QueryParams().Has("field") {
		componentSpecifiers := make([]choremodels.NewChoreTemplateComponent, 0)
		for i := range params.Fields {
			field := params.Fields[i]
			switch field {
			case string(choremodels.NewChoreTemplateComponentBaseInputs):
				componentSpecifiers = append(
					componentSpecifiers,
					choremodels.NewChoreTemplateComponentBaseInputs,
				)
			case string(choremodels.NewChoreTemplateComponentInputTypeSelector):
				componentSpecifiers = append(
					componentSpecifiers,
					choremodels.NewChoreTemplateComponentInputTypeSelector,
				)
			default:
				logging.Default().Info("invalid field specified", "field", field)
				return ctx.String(
					http.StatusUnprocessableEntity,
					"invalid field specified",
				)
			}
		}

		return h.templater.JoinedComponents(
			withNewChoreCtx(ctx, ctxValues),
			ctx.Response(),
			componentSpecifiers...)
	}

	return h.templater.NewChorePage(withNewChoreCtx(ctx, ctxValues), ctx.Response())
}

func withNewChoreCtx(
	ctx *echo.Context,
	values *choremodels.NewChoreCtxValues,
) context.Context {
	return choremodels.WithNewChoreCtxValues(ctx.Request().Context(), values)
}
