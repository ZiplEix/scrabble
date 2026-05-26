// api/middleware/accesslog/accesslog.go
package accesslog

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/ZiplEix/scrabble/api/middleware/logctx"
	"github.com/ZiplEix/scrabble/api/pkg/logger"
)

func isPreflight(c echo.Context) bool {
	r := c.Request()
	return r.Method == http.MethodOptions &&
		r.Header.Get("Origin") != "" &&
		r.Header.Get(echo.HeaderAccessControlRequestMethod) != ""
}

func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Read and store body for logging
			var bodyStr string
			if c.Request().Body != nil {
				bodyBytes, _ := io.ReadAll(c.Request().Body)
				bodyStr = string(bodyBytes)
				// Restore body for handler
				c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			start := time.Now()
			err := next(c)
			lat := time.Since(start)

			if isPreflight(c) {
				return err
			}

			req := c.Request()
			res := c.Response()

			status := res.Status
			var he *echo.HTTPError
			if errors.As(err, &he) && he != nil && he.Code > 0 {
				status = he.Code
			}

			fields := []any{
				"method", req.Method,
				"path", req.URL.Path,
				"status", status,
				"latency_ms", lat.Milliseconds(),
				"remote_ip", c.RealIP(),
				"user_agent", req.UserAgent(),
				"body", bodyStr,
			}

			for k, v := range logctx.All(c) {
				fields = append(fields, k, v)
			}
			if err != nil {
				fields = append(fields, "error", fmt.Sprint(err))
			}

			ctx := req.Context()
			switch {
			case status >= 500:
				logger.Error(ctx, "http_request", fields...)
			case status >= 400:
				logger.Warn(ctx, "http_request", fields...)
			default:
				logger.Info(ctx, "http_request", fields...)
			}
			return err
		}
	}
}
