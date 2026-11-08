package chores

import (
	"net/url"

	"github.com/labstack/echo/v5"
)

func collectionScopeHref(href string, mine bool) string {
	target, _ := url.Parse(href)
	query := target.Query()
	query.Del("offset")
	query.Del("owner_mode")
	query.Del("owner_id")
	query.Del("scope")
	if mine {
		query.Set("scope", "mine")
	}
	target.RawQuery = query.Encode()
	return target.RequestURI()
}

func withCollectionReturn(href, returnTo string) string {
	if returnTo == "" {
		return href
	}
	target, _ := url.Parse(href)
	query := target.Query()
	query.Set("return_to", returnTo)
	target.RawQuery = query.Encode()
	return target.RequestURI()
}

func collectionReturn(ctx *echo.Context) string {
	value := ctx.QueryParam("return_to")
	if value == "" {
		current, err := url.Parse(ctx.Request().Header.Get("HX-Current-URL"))
		if err == nil && current.Host == ctx.Request().Host {
			value = current.Query().Get("return_to")
		}
	}
	target, err := url.Parse(value)
	if err != nil || target.IsAbs() || target.Host != "" ||
		target.Path != choreCollectionHref {
		return ""
	}
	return target.RequestURI()
}
