package chores

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
	PatchStatus(id int, isComplete bool) error
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
		panic(err)
	}
	limitStr := ctx.QueryParamOr("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		panic(err)
	}
	content := ctx.QueryParamOr("content", "all")

	chores, err := h.persister.GetBatch(offset, limit)
	if err != nil {
		log.Println(err)
		return ctx.String(http.StatusInternalServerError, "something went wrong")
	}

	switch content {
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
	isCompleteStr := ctx.FormValueOr("complete", "")
	if isCompleteStr == "" {
		return ctx.String(http.StatusUnprocessableEntity, "complete missing")
	}
	err = h.persister.PatchStatus(id, isCompleteStr == "true")
	if err != nil {
		slog.Error("patch chore status error", "err", err)
		return ctx.String(http.StatusInternalServerError, "something went wrong")
	}

	ctx.Response().WriteHeader(http.StatusOK)

	return nil
}
