package middleware

import (
	"net/http"
)

// HTTPSOnly はhttpsOnlyがtrueの場合、HTTPリクエストをHTTPSにリダイレクトします
// HTTPS_ONLY環境変数によって制御されます
func HTTPSOnly(httpsOnly bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if httpsOnly && r.Header.Get("X-Forwarded-Proto") != "https" && r.TLS == nil {
				// HTTPSにリダイレクト
				target := "https://" + r.Host + r.RequestURI
				http.Redirect(w, r, target, http.StatusMovedPermanently)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
