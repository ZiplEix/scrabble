package logctx

import "github.com/labstack/echo/v4"

const key = "_logctx"

func Add(c echo.Context, k string, v any) {
	m := All(c)
	m[k] = v
	c.Set(key, m)
}

func Merge(c echo.Context, kv map[string]any) {
	m := All(c)
	for k, v := range kv {
		m[k] = v
	}
	c.Set(key, m)
}

func All(c echo.Context) map[string]any {
	if v := c.Get(key); v != nil {
		if m, ok := v.(map[string]any); ok {
			return m
		}
	}
	return map[string]any{}
}
