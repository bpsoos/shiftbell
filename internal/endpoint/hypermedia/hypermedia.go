package hypermedia

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v5"
)

const MediaType = "application/vnd.shiftbell+json"

type Link struct {
	Href string `json:"href"`
}

type ActionField struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Required  bool   `json:"required"`
	MaxLength int    `json:"max_length"`
}

type Action struct {
	Href        string        `json:"href"`
	Method      string        `json:"method"`
	ContentType string        `json:"content_type"`
	Fields      []ActionField `json:"fields"`
}

func Accepts(request *http.Request) bool {
	return request.Header.Get("Accept") == MediaType
}

func JSON(ctx *echo.Context, status int, value any) error {
	ctx.Response().Header().Set(echo.HeaderContentType, MediaType)
	ctx.Response().Header().Set("Vary", "Accept")
	ctx.Response().WriteHeader(status)
	return json.NewEncoder(ctx.Response()).Encode(value)
}
