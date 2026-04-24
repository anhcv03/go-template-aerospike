package common

import (
	"context"

	"github.com/labstack/echo/v4"
)

func AddToContext(c echo.Context) context.Context {
	newCtx := context.WithValue(c.Request().Context(), c.Request().Header.Get("X-Username"), "")
	newCtx = context.WithValue(newCtx, c.Request().Header.Get("X-Unitcode"), "")
	newCtx = context.WithValue(newCtx, c.Request().Header.Get("X-Unitname"), "")

	return newCtx
}
