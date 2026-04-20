package choretypes

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/bpsoos/shiftbell/internal/models"
	"github.com/labstack/echo/v5"
)

type Templater interface {
	PageWithLayout(
		context.Context,
		io.Writer,
		int,
		int,
		*models.GetChoreTypeBatchResult,
	) error
	Page(
		context.Context,
		io.Writer,
		int,
		int,
		*models.GetChoreTypeBatchResult,
	) error
	Table(
		context.Context,
		io.Writer,
		int,
		int,
		*models.GetChoreTypeBatchResult,
	) error
}

type ChoreTypePersister interface {
	Create(name string, description string) error
	Delete(id int) error
	GetBatch(offset int, limit int) (*models.GetChoreTypeBatchResult, error)
}

type HandlerDeps struct {
	Templater          Templater
	ChoreTypePersister ChoreTypePersister
}

type Handler struct {
	templater          Templater
	choreTypePersister ChoreTypePersister
}

func NewHandler(deps *HandlerDeps) *Handler {
	return &Handler{
		templater:          deps.Templater,
		choreTypePersister: deps.ChoreTypePersister,
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
	content := ctx.QueryParamOr("content", "all")

	chores, err := h.choreTypePersister.GetBatch(offset, limit)
	if err != nil {
		slog.Error("get chore type batch error", "err", err)
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

func (h *Handler) Delete(ctx *echo.Context) error {
	choreTypeIdStr := ctx.ParamOr("id", "")
	if choreTypeIdStr == "" {
		slog.Info("chore type id missing")
		return ctx.String(http.StatusUnprocessableEntity, "missing chore type id")
	}
	choreTypeId, err := strconv.Atoi(choreTypeIdStr)
	if err != nil {
		slog.Info("invalid chore type id")
		return ctx.String(http.StatusUnprocessableEntity, "invalid chore type id")
	}
	err = h.choreTypePersister.Delete(choreTypeId)
	if err != nil {
		slog.Info("chore type delete: %v", "err", err)
		return ctx.String(http.StatusInternalServerError, "something went wrong")
	}

	ctx.Response().Header().Set("HX-Trigger", "load-chore-types")
	ctx.Response().WriteHeader(http.StatusOK)
	return nil
}

func (h *Handler) Create(ctx *echo.Context) error {
	name := ctx.FormValue("name")
	if name == "" {
		slog.Info("missing name for create chore type")
		return ctx.String(http.StatusUnprocessableEntity, "missing name")
	}
	description := ctx.FormValue("description")

	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)

	err := h.choreTypePersister.Create(name, description)
	if err != nil {
		slog.Error("create chore type error", "err", err)
		return ctx.String(http.StatusInternalServerError, "something went wrong")
	}
	ctx.Response().Header().Set("HX-Trigger", "load-chore-types")

	return ctx.String(http.StatusOK, "created")
}
