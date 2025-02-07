package constants

type ContextKey string

const (
	XAmznTraceIDHeader            = "X-Amzn-Trace-Id"
	TraceIDKey         ContextKey = "traceID"
	UserContextKey     ContextKey = "auth-user"
)
