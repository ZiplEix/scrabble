package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestGetUserID_Found(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set(UserIDKey, int64(42))

	id, ok := GetUserID(c)
	if !ok {
		t.Fatalf("expected ok=true, got false")
	}
	if id != 42 {
		t.Fatalf("expected id=42, got %d", id)
	}
}

func TestGetUserID_NotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	id, ok := GetUserID(c)
	if ok {
		t.Fatalf("expected ok=false when key missing, got true (id=%d)", id)
	}
}

func TestGetUserID_WrongType(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set(UserIDKey, "42")
	_, ok := GetUserID(c)
	if ok {
		t.Fatalf("expected ok=false when value is wrong type, got true")
	}
}
