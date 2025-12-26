package middleware

import (
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CompactLogging returns a Gin middleware that outputs terminal/k9s-friendly one-line logs.
// Detailed information (stack traces, etc.) is only output on errors.
func CompactLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		// Basic one-line format: LEVEL [METHOD] /path STATUS latency
		baseLog := fmt.Sprintf("[%s] %s %d %dms", method, path, status, latency.Milliseconds())

		switch {
		case status >= 500:
			// 5xx: Output detailed error information
			logDetailedError(c, baseLog)
		case status >= 400:
			// 4xx: One line + brief reason
			reason := getErrorReason(c)
			if reason != "" {
				slog.Warn(baseLog + " -> " + reason)
			} else {
				slog.Warn(baseLog)
			}
		default:
			// 2xx, 3xx: One line only
			slog.Info(baseLog)
		}
	}
}

// logDetailedError outputs detailed information for 5xx errors.
func logDetailedError(c *gin.Context, baseLog string) {
	var details []string

	// Error message
	if err, exists := c.Get("error"); exists {
		details = append(details, fmt.Sprintf("error: %v", err))
	}

	// User ID
	if userID, exists := c.Get("user_id"); exists {
		details = append(details, fmt.Sprintf("user_id: %v", userID))
	}

	// Request ID
	if reqID := GetRequestID(c); reqID != "" {
		details = append(details, fmt.Sprintf("request_id: %s", reqID))
	}

	// Stack trace (simplified)
	stack := getSimplifiedStack(3, 5) // skip 3 frames, get 5 frames
	if len(stack) > 0 {
		details = append(details, "stack:")
		for _, frame := range stack {
			details = append(details, "    "+frame)
		}
	}

	// Output
	slog.Error(baseLog)
	for _, detail := range details {
		fmt.Printf("      -> %s\n", detail)
	}
}

// getErrorReason retrieves the error reason from the context.
func getErrorReason(c *gin.Context) string {
	if reason, exists := c.Get("error_reason"); exists {
		return fmt.Sprintf("%v", reason)
	}
	return ""
}

// getSimplifiedStack returns a simplified stack trace.
func getSimplifiedStack(skip, count int) []string {
	var frames []string
	pcs := make([]uintptr, count)
	n := runtime.Callers(skip, pcs)

	for i := 0; i < n && i < count; i++ {
		fn := runtime.FuncForPC(pcs[i])
		if fn == nil {
			continue
		}
		file, line := fn.FileLine(pcs[i])
		// Extract filename only
		parts := strings.Split(file, "/")
		shortFile := parts[len(parts)-1]
		// Simplify function name
		funcName := fn.Name()
		funcParts := strings.Split(funcName, ".")
		shortFunc := funcParts[len(funcParts)-1]

		frames = append(frames, fmt.Sprintf("%s:%d %s()", shortFile, line, shortFunc))
	}
	return frames
}

// JSONLogging returns a Gin middleware for production JSON logging.
func JSONLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		attrs := []any{
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"ip", c.ClientIP(),
			"request_id", GetRequestID(c),
		}

		if userID, exists := c.Get("user_id"); exists {
			attrs = append(attrs, "user_id", userID)
		}

		if err, exists := c.Get("error"); exists {
			attrs = append(attrs, "error", err)
		}

		switch {
		case status >= 500:
			slog.Error("request", attrs...)
		case status >= 400:
			slog.Warn("request", attrs...)
		default:
			slog.Info("request", attrs...)
		}
	}
}

// NewLoggingMiddleware returns the appropriate logging middleware based on environment.
func NewLoggingMiddleware(env string) gin.HandlerFunc {
	switch env {
	case "production":
		return JSONLogging()
	default:
		return CompactLogging()
	}
}
