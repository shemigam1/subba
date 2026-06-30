// Package render centralizes JSON responses and the error envelope, so every error
// carries a request_id that matches the structured logs.
package render

import (
	"github.com/gin-gonic/gin"
)

const RequestIDKey = "request_id"

// Error is the shared error envelope (see Error schema in openapi.yaml).
type Error struct {
	Code      string            `json:"error"`
	Message   string            `json:"message,omitempty"`
	RequestID string            `json:"request_id"`
	Fields    map[string]string `json:"fields,omitempty"`
}

// RequestID returns the correlation id set by the RequestID middleware.
func RequestID(c *gin.Context) string {
	if v, ok := c.Get(RequestIDKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// JSON writes a success payload.
func JSON(c *gin.Context, status int, body any) {
	c.JSON(status, body)
}

// Err writes the error envelope and aborts the handler chain.
func Err(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, Error{Code: code, Message: message, RequestID: RequestID(c)})
}

// ValidationErr writes a 422 with per-field messages.
func ValidationErr(c *gin.Context, fields map[string]string) {
	c.AbortWithStatusJSON(422, Error{Code: "validation_error", Message: "invalid input", RequestID: RequestID(c), Fields: fields})
}
