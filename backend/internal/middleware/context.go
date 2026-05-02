package middleware

import "context"

const RequestIDKey ContextKey = "requestID"

func RequestIDFrom(ctx context.Context) string {
	if v, ok := ctx.Value(RequestIDKey).(string); ok {
		return v
	}
	return ""
}
