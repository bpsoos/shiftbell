package endpoint

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/bpsoos/shiftbell/internal/models"
	"github.com/labstack/echo/v5"
)

type Templater interface {
	Home(int, int, *models.GetChoreTypeBatchResult, context.Context, io.Writer) error
	GetChoreBatch(int, int, *models.GetChoreTypeBatchResult, context.Context, io.Writer) error
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

func (h *Handler) Home(ctx *echo.Context) error {
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
		ctx.Response().Header().Set("HX-Push-Url", fmt.Sprintf("?offset=%d&limit=%d", offset, limit))
		ctx.Response().WriteHeader(http.StatusOK)

		return h.templater.GetChoreBatch(offset, limit, chores, ctx.Request().Context(), ctx.Response())
	case "all":
		ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		ctx.Response().Header().Set("HX-Push-Url", fmt.Sprintf("?offset=%d&limit=%d", offset, limit))
		ctx.Response().WriteHeader(http.StatusOK)

		return h.templater.Home(offset, limit, chores, ctx.Request().Context(), ctx.Response())
	default:
		slog.Error("unknown conent", "content", content)
		return ctx.String(http.StatusUnprocessableEntity, "unknown content")
	}
}

func (h *Handler) ViewSettings(ctx *echo.Context) error {
	offsetStr := ctx.QueryParamOr("offset", "0")
	limitStr := ctx.QueryParamOr("limit", "10")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		panic(err)
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		panic(err)
	}
	ctx.Response().Header().Set("HX-Push-Url", fmt.Sprintf("?offset=%d&limit=%d", offset, limit))
	ctx.Response().Header().Set("HX-Target", "none")
	return ctx.String(http.StatusOK, "OK")
}

func (h *Handler) GetChoreBatch(ctx *echo.Context) error {
	offsetStr := ctx.QueryParamOr("offset", "0")
	limitStr := ctx.QueryParamOr("limit", "10")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		panic(err)
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		panic(err)
	}

	chores, err := h.choreTypePersister.GetBatch(offset, limit)
	if err != nil {
		log.Println(err)
		return ctx.String(http.StatusInternalServerError, "something went wrong")
	}

	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	ctx.Response().Header().Set("HX-Push-Url", fmt.Sprintf("?offset=%d&limit=%d", offset, limit))
	ctx.Response().WriteHeader(http.StatusOK)
	return h.templater.GetChoreBatch(offset, limit, chores, ctx.Request().Context(), ctx.Response())
}

func (h *Handler) CreateChore(ctx *echo.Context) error {
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
	ctx.Response().Header().Set("HX-Trigger", "load-chores")

	return ctx.String(http.StatusOK, "created")
}
