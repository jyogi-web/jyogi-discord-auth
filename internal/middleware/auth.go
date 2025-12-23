package middleware

import (
	"context"
	"net/http"

	"github.com/jyogi-web/jyogi-discord-auth/pkg/jwt"
)

// contextKey はコンテキストキーの型です
type contextKey string

const (
	// UserClaimsKey はユーザークレームのコンテキストキーです
	UserClaimsKey contextKey = "user_claims"
)

// JWTAuth はJWT認証ミドルウェアを返します
// jwtSecret: JWT検証用のシークレットキー
func JWTAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Authorization ヘッダーからトークンを取得
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeJSONError(w, http.StatusUnauthorized, "missing_token", "Authorization header is required")
				return
			}

			// Bearer トークンの形式を確認
			var tokenString string
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				tokenString = authHeader[7:]
			} else {
				writeJSONError(w, http.StatusUnauthorized, "invalid_token_format", "Authorization header must be in 'Bearer <token>' format")
				return
			}

			// トークンを検証
			claims, err := jwt.ValidateToken(tokenString, jwtSecret)
			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, "invalid_token", "Token is invalid or expired")
				return
			}

			// クレームをコンテキストに追加
			ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// writeJSONError はJSON形式のエラーレスポンスを書き込みます
func writeJSONError(w http.ResponseWriter, status int, errorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"error":"` + errorCode + `","message":"` + message + `"}`))
}

// GetUserClaims はコンテキストからユーザークレームを取得します
func GetUserClaims(ctx context.Context) (*jwt.Claims, bool) {
	claims, ok := ctx.Value(UserClaimsKey).(*jwt.Claims)
	return claims, ok
}
