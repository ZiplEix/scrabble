package requestid

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/labstack/echo/v4"
)

const (
	// Header sous lequel on renvoie l'ID au client
	DefaultHeaderName = echo.HeaderXRequestID // "X-Request-ID"
	// Clé de contexte pour retrouver l'ID dans les handlers/middlewares
	ContextKey = "request_id"
)

// Middleware renvoie un echo.MiddlewareFunc qui :
// - lit un ID entrant éventuel (X-Request-ID ou X-Correlation-ID),
// - sinon génère un ID cryptographiquement aléatoire,
// - stocke l'ID dans le contexte,
// - et l'ajoute à la réponse (header X-Request-ID).
func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			id := incomingID(c)
			if id == "" {
				id = newID()
			}
			// Expose dans le contexte pour les handlers & logs
			c.Set(ContextKey, id)

			// S'assure que le header est bien présent sur *toutes* les réponses,
			// y compris en cas d'erreur, juste avant l'écriture des headers.
			c.Response().Before(func() {
				c.Response().Header().Set(DefaultHeaderName, id)
			})

			return next(c)
		}
	}
}

// Get récupère l'ID depuis le contexte.
func Get(c echo.Context) string {
	v := c.Get(ContextKey)
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// incomingID tente de récupérer un ID fourni par le client.
func incomingID(c echo.Context) string {
	h := c.Request().Header
	if v := strings.TrimSpace(h.Get(echo.HeaderXRequestID)); v != "" {
		return v
	}
	// Support optionnel de X-Correlation-ID si un proxy/edge le pose
	if v := strings.TrimSpace(h.Get("X-Correlation-ID")); v != "" {
		return v
	}
	return ""
}

// newID génère un ID hex(16octets) => 32 chars, collision très improbable.
func newID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
