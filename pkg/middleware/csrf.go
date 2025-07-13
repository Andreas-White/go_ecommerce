package middleware

import (
	"net/http"
)

// CSRFMiddleware checks CSRF token for state-changing requests
func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only check CSRF for state-changing methods
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		csrfCookie, err := r.Cookie("csrf_token")
		if err != nil || csrfCookie.Value == "" {
			http.Error(w, "Missing CSRF cookie", http.StatusForbidden)
			return
		}
		csrfHeader := r.Header.Get("X-CSRF-Token")
		if csrfHeader == "" {
			http.Error(w, "Missing CSRF header", http.StatusForbidden)
			return
		}
		if csrfCookie.Value != csrfHeader {
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
