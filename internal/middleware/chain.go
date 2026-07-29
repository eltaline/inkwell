package middleware

import "net/http"

// Chain composes multiple middleware into a single middleware.
// Middleware is applied in the order given, so the first middleware
// in the list is the outermost (executed first on request, last on response).
func Chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}
