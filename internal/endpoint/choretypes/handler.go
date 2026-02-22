package choretypes

import (
	"context"
	"io"
	"log"
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
	Create(description string, intervalDays int) error
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
		panic(err)
	}
	limitStr := ctx.QueryParamOr("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		panic(err)
	}
	content := ctx.QueryParamOr("content", "all")

	chores, err := h.choreTypePersister.GetBatch(offset, limit)
	if err != nil {
		log.Println(err)
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
		slog.Error("unknown conent", "content", content)
		return ctx.String(http.StatusUnprocessableEntity, "unknown content")
	}
}

func (h *Handler) Create(ctx *echo.Context) error {
	description := ctx.FormValue("description")
	intervalDaysStr := ctx.FormValue("interval-days")
	intervalDays, err := strconv.Atoi(intervalDaysStr)
	if err != nil {
		panic(err)
	}

	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)

	err = h.choreTypePersister.Create(description, intervalDays)
	if err != nil {
		log.Println(err)
		return ctx.String(http.StatusInternalServerError, "something went wrong")
	}
	ctx.Response().Header().Set("HX-Trigger", "load-chore-types")

	return ctx.String(http.StatusOK, "created")
}
