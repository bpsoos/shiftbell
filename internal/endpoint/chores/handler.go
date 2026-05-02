package chores

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/bpsoos/shiftbell/internal/models"
	"github.com/labstack/echo/v5"
)

type Templater interface {
	Table(
		context.Context,
		io.Writer,
		int,
		int,
		*models.GetChoreBatchResult,
	) error
	Chore(
		context.Context,
		io.Writer,
		*models.Chore,
	) error
	ChoreForEdit(
		context.Context,
		io.Writer,
		*models.Chore,
	) error
	Page(
		context.Context,
		io.Writer,
		int,
		int,
		*models.GetChoreBatchResult,
		*models.GetChoreTypeBatchResult,
		*models.ChoreType,
	) error
	PageWithLayout(
		context.Context,
		io.Writer,
		int,
		int,
		*models.GetChoreBatchResult,
	) error
	NewChoreByTypePage(
		context.Context,
		io.Writer,
		*models.ChoreType,
	) error
	NewManualChorePage(
		context.Context,
		io.Writer,
	) error
}

type ChoreTypePersister interface {
	Get(id int) (*models.ChoreType, error)
}

type Persister interface {
	GetBatch(offset int, limit int) (*models.GetChoreBatchResult, error)
	Get(id int) (*models.Chore, error)
	MarkComplete(id int, completedAt time.Time) error
	SetLastCompletedAt(id int, lastCompletedAt time.Time) (*models.Chore, error)
}

type HandlerDeps struct {
	Templater          Templater
	Persister          Persister
	ChoreTypePersister ChoreTypePersister
}

type Handler struct {
	templater          Templater
	persister          Persister
	choreTypePersister ChoreTypePersister
}

func NewHandler(deps *HandlerDeps) *Handler {
	return &Handler{
		templater:          deps.Templater,
		persister:          deps.Persister,
		choreTypePersister: deps.ChoreTypePersister,
	}
}

func (h *Handler) Get(ctx *echo.Context) error {
	idStr := ctx.ParamOr("id", "")
	if idStr == "" {
		return ctx.String(http.StatusUnprocessableEntity, "id missing")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Info("invalid id received", "err", err)
		return ctx.String(http.StatusUnprocessableEntity, "invalid id")
	}

	chore, err := h.persister.Get(id)
	if err != nil {
		slog.Error("get batch error", "err", err)
		return ctx.String(http.StatusInternalServerError, "something went wrong")
	}

	getForType := ctx.QueryParamOr("for", "readonly")
	switch getForType {
	case "readonly":
		ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		ctx.Response().WriteHeader(http.StatusOK)

		return h.templater.Chore(ctx.Request().Context(), ctx.Response(), chore)
	case "edit":
		ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		ctx.Response().WriteHeader(http.StatusOK)

		return h.templater.ChoreForEdit(ctx.Request().Context(), ctx.Response(), chore)
	default:
		slog.Error("unknown get for cause", "get_for_type", getForType)
		return ctx.String(http.StatusUnprocessableEntity, "unknown get for type")
	}

}

func (h *Handler) GetBatch(ctx *echo.Context) error {
	offsetStr := ctx.QueryParamOr("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		slog.Info("invalid offset", "err", err)
		return ctx.String(http.StatusUnprocessableEntity, "invalid offset")
	}
	limitStr := ctx.QueryParamOr("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		slog.Info("invalid limit", "err", err)
		return ctx.String(http.StatusUnprocessableEntity, "invalid limit")
	}
	statusFilter := ctx.QueryParamOr("status", "incomplete")
	if statusFilter != "incomplete" {
		slog.Info("unsupported status filter", "status_filter", statusFilter)
		return ctx.String(http.StatusUnprocessableEntity, "unsupported status filter")
	}
	content := ctx.QueryParamOr("content", "all")

	chores, err := h.persister.GetBatch(offset, limit)
	if err != nil {
		slog.Error("get batch error", "err", err)
		return ctx.String(http.StatusInternalServerError, "something went wrong")
	}

	switch content {
	case "table":
		ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		ctx.Response().WriteHeader(http.StatusOK)

		return h.templater.Table(ctx.Request().Context(), ctx.Response(), offset, limit, chores)
	case "all":
		ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		ctx.Response().WriteHeader(http.StatusOK)

		if ctx.Request().Header.Get("HX-Request") == "true" {
			return h.templater.Page(ctx.Request().Context(), ctx.Response(), offset, limit, chores, nil, nil)
		}

		return h.templater.PageWithLayout(ctx.Request().Context(), ctx.Response(), offset, limit, chores)
	default:
		slog.Error("unknown content", "content", content)
		return ctx.String(http.StatusUnprocessableEntity, "unknown content")
	}
}

func (h *Handler) Patch(ctx *echo.Context) error {
	idStr := ctx.ParamOr("id", "")
	if idStr == "" {
		return ctx.String(http.StatusUnprocessableEntity, "id missing")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Info("invalid id received", "err", err)
		return ctx.String(http.StatusUnprocessableEntity, "invalid id")
	}
	status := ctx.FormValueOr("status", "")
	switch status {
	case "":
	case "complete":
	case "incomplete":
		return ctx.String(http.StatusConflict, "setting status to incomplete dissalowed")
	default:
		slog.Info("unknown status", "status", status)
		return ctx.String(http.StatusUnprocessableEntity, "unknown status")
	}

	lastCompletedAtStr := ctx.FormValueOr("lastCompletedAt", "")

	if lastCompletedAtStr == "" && status == "" {
		slog.Info("patch request request content empty")
		return ctx.String(http.StatusUnprocessableEntity, "patch content missing")
	}
	slog.Info("updating lastCompletedAt", "last_completed_at", lastCompletedAtStr, "status", status)

	var lastCompletedAt time.Time
	if lastCompletedAtStr != "" {
		lastCompletedAt, err = time.Parse("2006-01-02", lastCompletedAtStr)
		if err != nil {
			slog.Info("invalid lastCompletedAt", "last_completed_at", lastCompletedAtStr)
			return ctx.String(http.StatusUnprocessableEntity, "invalid lastCompletedAt")
		}
	}

	if status != "" {
		err = h.persister.MarkComplete(id, time.Now())
		if err != nil {
			slog.Error("patch chore status error", "err", err)
			return ctx.String(http.StatusInternalServerError, "something went wrong")
		}

		ctx.Response().Header().Set("HX-Trigger", "load-chores")
	}
	if lastCompletedAtStr != "" {
		slog.Info("updating lastCompletedAt", "last_completed_at", lastCompletedAtStr)
		chore, err := h.persister.SetLastCompletedAt(id, lastCompletedAt)
		if err != nil {
			slog.Error("set last updated at error", "err", err)
			return ctx.String(http.StatusInternalServerError, "something went wrong")
		}
		ctx.Response().Header().Set("HX-Trigger", "load-chores")
		ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		ctx.Response().WriteHeader(http.StatusOK)

		return h.templater.Chore(ctx.Request().Context(), ctx.Response(), chore)
	}

	return ctx.String(http.StatusOK, "OK")
}

func (h *Handler) New(ctx *echo.Context) error {
	var selectedChoreType *models.ChoreType
	choreTypeIdStr := ctx.QueryParamOr("choreTypeId", "")
	if choreTypeIdStr != "" {
		selectedChoreTypeId, err := strconv.Atoi(choreTypeIdStr)
		if err != nil {
			slog.Info("chore type id parse error", "err", err)
			return ctx.String(http.StatusUnprocessableEntity, "invalid chore type id")
		}
		selectedChoreType, err = h.choreTypePersister.Get(selectedChoreTypeId)
		if err != nil {
			slog.Info("get chore type error", "err", err)
			return ctx.String(http.StatusInternalServerError, "something went wrong")
		}
	}

	inputType := ctx.QueryParamOr("inputType", "selectChoreType")
	switch inputType {
	case "selectChoreType":
		return h.templater.NewChoreByTypePage(ctx.Request().Context(), ctx.Response(), selectedChoreType)
	case "manual":
		return h.templater.NewManualChorePage(ctx.Request().Context(), ctx.Response())
	default:
		slog.Info("invalid input type", "input_type", inputType)
		return ctx.String(http.StatusUnprocessableEntity, "invalid input type")
	}
}
