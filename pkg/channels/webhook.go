package channels

import "net/http"

// WebhookHandler is an optional interface for channels that receive messages
// via HTTP webhooks. Manager discovers channels implementing this interface
// and registers them on the shared HTTP server.
type WebhookHandler interface {
	// WebhookPath returns the path to mount this handler on the shared server.
	// Examples: "/webhook/line", "/webhook/wecom"
	WebhookPath() string
	http.Handler // ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// HealthChecker is an optional interface for channels that expose
// a health check endpoint on the shared HTTP server.
type HealthChecker interface {
	HealthPath() string
	HealthHandler(w http.ResponseWriter, r *http.Request)
}

// LiveHealthChecker is an optional interface for channels that contribute a
// live readiness signal to the shared /ready probe. Manager registers the
// HealthCheck callback with the health server so each /ready request causes
// it to be evaluated. Use this for channels whose external connection
// (e.g., Discord WebSocket) can silently fail while the gateway process is
// still running — the documented symptom is that /health stays green while
// the bot stops responding. The function should be cheap; expensive probes
// belong in a background goroutine that flips a cached state.
type LiveHealthChecker interface {
	// HealthCheck returns (ok, detail). When ok is false, /ready returns
	// 503 with detail in the per-check Message field.
	HealthCheck() (bool, string)
}
