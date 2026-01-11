package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/jyogi-web/jyogi-discord-auth/internal/service"
)

// SessionAuth returns a middleware that validates session cookies.
// It redirects to /auth/login for browser requests and returns 401 for API requests.
func SessionAuth(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get session token from cookie
			cookie, err := r.Cookie("session_token")
			var token string
			if err == nil {
				token = cookie.Value
			}

			// Validate session if token exists
			// We don't necessarily need the user object here if we are just checking validity,
			// but GetUserBySessionToken is the method to check validity.
			if token != "" {
				_, err = authService.GetUserBySessionToken(r.Context(), token)
			}

			// If no token or invalid session (err is from GetUserBySessionToken)
			if token == "" || err != nil {
				// Check if it's an API request (JSON) or DELETE method
				if r.Method == http.MethodDelete || strings.Contains(r.Header.Get("Accept"), "application/json") {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`{"error":"unauthorized","message":"Authentication required"}`))
					return
				}

				// Otherwise redirect to login
				// We preserve the redirect_uri but ensure it is safe (relative path only)
				redirectURI := r.URL.Path
				if r.URL.RawQuery != "" {
					redirectURI += "?" + r.URL.RawQuery
				}

				// Validate redirectURI to prevent open redirects
				// It must start with / and parsing it should not result in an absolute URL (Scheme/Host must be empty)
				parsedURL, err := url.Parse(redirectURI)
				if err != nil || parsedURL.Scheme != "" || parsedURL.Host != "" || !strings.HasPrefix(redirectURI, "/") {
					redirectURI = "/"
				}

				http.Redirect(w, r, "/auth/login?redirect_uri="+url.QueryEscape(redirectURI), http.StatusFound)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
