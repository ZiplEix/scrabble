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
	"go.uber.org/zap"

	"github.com/ZiplEix/scrabble/api/middleware/logctx"
	reqid "github.com/ZiplEix/scrabble/api/middleware/request_id"
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

			fields := []zap.Field{
				zap.String("request_id", reqid.Get(c)),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path),
				zap.Int("status", status),
				zap.Int64("latency_ms", lat.Milliseconds()),
				zap.String("remote_ip", c.RealIP()),
				zap.String("user_agent", req.UserAgent()),
				zap.String("body", bodyStr),
			}

			for k, v := range logctx.All(c) {
				fields = append(fields, zap.Any(k, v))
			}
			if err != nil {
				fields = append(fields, zap.String("error", fmt.Sprint(err)))
			}

			switch {
			case status >= 500:
				zap.L().Error("http_request", fields...)
			case status >= 400:
				zap.L().Warn("http_request", fields...)
			default:
				zap.L().Info("http_request", fields...)
			}
			return err
		}
	}
}
