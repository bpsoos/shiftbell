package choretemplates

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

const (
	flashCookieName               = "shiftbell_template_flash"
	templateDeactivatedFlashValue = "template-deactivated"
	templateDeactivatedNotice     = "Template deactivated."
)

func setFlashCookie(ctx *echo.Context, value string) {
	cookie := &http.Cookie{
		Name:     flashCookieName,
		Value:    value,
		Path:     "/chore-templates",
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
	if cookie.Value == templateDeactivatedFlashValue {
		return templateDeactivatedNotice
	}
	return ""
}

func clearFlashCookie(ctx *echo.Context) {
	cookie := &http.Cookie{
		Name:     flashCookieName,
		Path:     "/chore-templates",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	}
	ctx.Response().Header().Add(echo.HeaderSetCookie, cookie.String())
}
