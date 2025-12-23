package middleware

import (
	"log"
	"net/http"
	"time"
)

// responseWriter はステータスコードをキャプチャするためにhttp.ResponseWriterをラップします
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logging はHTTPリクエストをメソッド、パス、ステータス、実行時間とともにログに記録します
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// ステータスコードをキャプチャするためにレスポンスライターをラップ
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// 次のハンドラーを呼び出し
		next.ServeHTTP(wrapped, r)

		// リクエストをログに記録
		duration := time.Since(start)
		log.Printf("[%s] %s %s - %d (%v)",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			wrapped.statusCode,
			duration,
		)
	})
}
