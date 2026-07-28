package choretemplates

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/bpsoos/shiftbell/internal/logging"
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	"github.com/labstack/echo/v5"
)

type Templater interface {
	PageWithLayout(
		context.Context,
		io.Writer,
		int,
		int,
		*models.GetChoreTemplateBatchResult,
	) error
	Page(
		context.Context,
		io.Writer,
		int,
		int,
		*models.GetChoreTemplateBatchResult,
	) error
	Table(
		context.Context,
		io.Writer,
		int,
		int,
		*models.GetChoreTemplateBatchResult,
	) error
	Selector(
		context.Context,
		io.Writer,
		int,
		int,
		*models.GetChoreTemplateBatchResult,
	) error
}

type ChoreTemplatePersister interface {
	Create(name string, description string) error
	Delete(id int) error
	GetBatch(offset int, limit int) (*models.GetChoreTemplateBatchResult, error)
}

type HandlerDeps struct {
	Templater              Templater
	ChoreTemplatePersister ChoreTemplatePersister
}

type Handler struct {
	templater              Templater
	choreTemplatePersister ChoreTemplatePersister
}

func NewHandler(deps *HandlerDeps) *Handler {
	return &Handler{
		templater:              deps.Templater,
		choreTemplatePersister: deps.ChoreTemplatePersister,
	}
}

func (h *Handler) GetBatch(ctx *echo.Context) error {
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
	content := ctx.QueryParamOr("content", "all")

	choreTemplates, err := h.choreTemplatePersister.GetBatch(offset, limit)
	if err != nil {
		logging.Default().Error("get chore template batch error", "err", err)
		return ctx.String(http.StatusInternalServerError, "something went wrong")
	}

	switch content {
	case "selector":
		ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		ctx.Response().WriteHeader(http.StatusOK)

		return h.templater.Selector(ctx.Request().Context(), ctx.Response(), offset, limit, choreTemplates)
	case "table":
		ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		ctx.Response().WriteHeader(http.StatusOK)

		return h.templater.Table(ctx.Request().Context(), ctx.Response(), offset, limit, choreTemplates)
	case "all":
		ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		ctx.Response().WriteHeader(http.StatusOK)

		if ctx.Request().Header.Get("HX-Request") == "true" {
			return h.templater.Page(ctx.Request().Context(), ctx.Response(), offset, limit, choreTemplates)
		}

		return h.templater.PageWithLayout(ctx.Request().Context(), ctx.Response(), offset, limit, choreTemplates)

	default:
		logging.Default().Error("unknown content", "content", content)
		return ctx.String(http.StatusUnprocessableEntity, "unknown content")
	}
}

func (h *Handler) Delete(ctx *echo.Context) error {
	choreTemplateIdStr := ctx.ParamOr("id", "")
	if choreTemplateIdStr == "" {
		logging.Default().Info("chore template id missing")
		return ctx.String(http.StatusUnprocessableEntity, "missing chore template id")
	}
	choreTemplateId, err := strconv.Atoi(choreTemplateIdStr)
	if err != nil {
		logging.Default().Info("invalid chore template id")
		return ctx.String(http.StatusUnprocessableEntity, "invalid chore template id")
	}
	err = h.choreTemplatePersister.Delete(choreTemplateId)
	if err != nil {
		logging.Default().Info("chore template delete", "err", err)
		return ctx.String(http.StatusInternalServerError, "something went wrong")
	}

	ctx.Response().Header().Set("HX-Trigger", "load-chore-templates")
	ctx.Response().WriteHeader(http.StatusOK)
	return nil
}

func (h *Handler) Create(ctx *echo.Context) error {
	name := ctx.FormValue("name")
	if name == "" {
		logging.Default().Info("missing name for create chore template")
		return ctx.String(http.StatusUnprocessableEntity, "missing name")
	}
	description := ctx.FormValue("description")

	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)

	err := h.choreTemplatePersister.Create(name, description)
	if err != nil {
		logging.Default().Error("create chore template error", "err", err)
		return ctx.String(http.StatusInternalServerError, "something went wrong")
	}
	ctx.Response().Header().Set("HX-Trigger", "load-chore-templates")

	return ctx.String(http.StatusOK, "created")
}
