package chores

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

const (
	flashCookieName                   = "shiftbell_flash"
	choreAndTemplateCreatedFlashValue = "chore-and-template-created"
	choreAndTemplateCreatedNotice     = "Chore added and template saved."
)

func setFlashCookie(ctx *echo.Context, value string) {
	cookie := &http.Cookie{
		Name:     flashCookieName,
		Value:    value,
		Path:     "/chores",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	ctx.Response().Header().Add(echo.HeaderSetCookie, cookie.String())
}

func consumeFlashCookie(ctx *echo.Context) string {
	cookie, err := ctx.Request().Cookie(flashCookieName)
	if err != nil {
		return ""
	}
	clearFlashCookie(ctx)
	if cookie.Value == choreAndTemplateCreatedFlashValue {
		return choreAndTemplateCreatedNotice
	}
	return ""
}

func clearFlashCookie(ctx *echo.Context) {
	cookie := &http.Cookie{
		Name:     flashCookieName,
		Path:     "/chores",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	}
	ctx.Response().Header().Add(echo.HeaderSetCookie, cookie.String())
}
