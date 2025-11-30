package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

const (
	// MaxRequestBodySize limits request body size to 1MB
	MaxRequestBodySize = 1 << 20 // 1 MB
	contentTypeJSON    = "application/json"
)

// ValidateJSONContentType validates that POST/PUT/PATCH requests have JSON Content-Type.
func ValidateJSONContentType(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only validate Content-Type for methods that typically have a body
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			contentType := r.Header.Get("Content-Type")
			if contentType == "" {
				slog.Warn("missing Content-Type header",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
				)
				http.Error(w, "Content-Type header is required", http.StatusBadRequest)
				return
			}

			// Allow Content-Type with charset, e.g., "application/json; charset=utf-8"
			if !strings.HasPrefix(contentType, contentTypeJSON) {
				slog.Warn("invalid Content-Type header",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("content_type", contentType),
				)
				http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
				return
			}
		}

		next(w, r)
	}
}

// ValidateBodySize limits the size of request body to prevent DoS attacks.
func ValidateBodySize(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Limit body size
		r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBodySize)

		next(w, r)
	}
}

// ValidateJSONStructure validates that request body is valid JSON (without decoding into specific struct).
func ValidateJSONStructure(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only validate JSON for methods that have a body
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			// Read body to validate JSON structure
			body, err := io.ReadAll(r.Body)
			if err != nil {
				if err.Error() == "http: request body too large" {
					slog.Warn("request body too large",
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
					)
					http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
					return
				}
				slog.Warn("failed to read request body",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.Any("error", err),
				)
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}

			// Restore body for next handler
			r.Body = io.NopCloser(bytes.NewReader(body))

			// Validate JSON structure
			if len(body) > 0 {
				var jsonValue interface{}
				if err := json.Unmarshal(body, &jsonValue); err != nil {
					slog.Warn("invalid JSON structure",
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
						slog.Any("error", err),
					)
					http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
					return
				}
			}
		}

		next(w, r)
	}
}

// Chain combines multiple middleware functions into a single middleware.
func Chain(middlewares ...func(http.HandlerFunc) http.HandlerFunc) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}
