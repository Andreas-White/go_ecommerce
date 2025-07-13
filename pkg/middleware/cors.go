package middleware

import (
	"net/http"
)

// CORS middleware to handle cross-origin requests
func CORS(next http.Handler) http.Handler {
	// Define your trusted origins here
	trustedOrigins := map[string]bool{
		"http://localhost:3000": true,
		// "https://yourdomain.com": true,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Validate origin and set appropriate headers
		if trustedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			// For untrusted origins, don't set Allow-Origin header
			// This prevents the browser from making the request
			w.Header().Set("Access-Control-Allow-Credentials", "false")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-CSRF-Token")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
