package middleware

import "github.com/labstack/echo/v4"

func DisableCache() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Add("Cache-Control", "no-store")
			return next(c)
		}
	}
}
