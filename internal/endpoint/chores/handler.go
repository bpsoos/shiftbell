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
	Page(
		context.Context,
		io.Writer,
		int,
		int,
		*models.GetChoreBatchResult,
	) error
	PageWithLayout(
		context.Context,
		io.Writer,
		int,
		int,
		*models.GetChoreBatchResult,
	) error
}

type Persister interface {
	GetBatch(offset int, limit int) (*models.GetChoreBatchResult, error)
	MarkComplete(id int, completedAt time.Time) error
}

type HandlerDeps struct {
	Templater Templater
	Persister Persister
}

type Handler struct {
	templater Templater
	persister Persister
}

func NewHandler(deps *HandlerDeps) *Handler {
	return &Handler{
		templater: deps.Templater,
		persister: deps.Persister,
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
			return h.templater.Page(ctx.Request().Context(), ctx.Response(), offset, limit, chores)
		}

		return h.templater.PageWithLayout(ctx.Request().Context(), ctx.Response(), offset, limit, chores)
	default:
		slog.Error("unknown content", "content", content)
		return ctx.String(http.StatusUnprocessableEntity, "unknown content")
	}
}

func (h *Handler) PatchStatus(ctx *echo.Context) error {
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
	if status == "" {
		slog.Info("status missing")
		return ctx.String(http.StatusUnprocessableEntity, "status missing")
	}
	switch status {
	case "complete":
	case "incomplete":
		return ctx.String(http.StatusConflict, "setting status to incomplete dissalowed")
	default:
		slog.Info("unknown status", "status", status)
		return ctx.String(http.StatusUnprocessableEntity, "unknown status")
	}

	err = h.persister.MarkComplete(id, time.Now())
	if err != nil {
		slog.Error("patch chore status error", "err", err)
		return ctx.String(http.StatusInternalServerError, "something went wrong")
	}

	ctx.Response().Header().Set("HX-Trigger", "load-chores")
	ctx.Response().WriteHeader(http.StatusOK)

	return nil
}
