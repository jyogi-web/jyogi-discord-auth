package handler

import (
	"net/http"
	"time"

	"github.com/jyogi-web/jyogi-discord-auth/internal/service"
	"github.com/jyogi-web/jyogi-discord-auth/pkg/jwt"
)

// TokenHandler はトークンハンドラーを表します
type TokenHandler struct {
	authService *service.AuthService
	jwtSecret   string
}

// NewTokenHandler は新しいトークンハンドラーを作成します
func NewTokenHandler(authService *service.AuthService, jwtSecret string) *TokenHandler {
	return &TokenHandler{
		authService: authService,
		jwtSecret:   jwtSecret,
	}
}

// HandleIssueToken はセッショントークンからJWTを発行します
// POST /token
func (h *TokenHandler) HandleIssueToken(w http.ResponseWriter, r *http.Request) {
	// セッショントークンを取得
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "No active session")
		return
	}

	// セッショントークンからユーザー情報を取得
	user, err := h.authService.GetUserBySessionToken(r.Context(), sessionCookie.Value)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "invalid_session", "Session is invalid or expired")
		return
	}

	// JWTを生成（7日間有効）
	accessToken, err := jwt.GenerateToken(
		user.ID,
		user.DiscordID,
		user.Username,
		h.jwtSecret,
		7*24*time.Hour,
	)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "token_generation_failed", "Failed to generate access token")
		return
	}

	// レスポンスを返す
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   7 * 24 * 3600, // 7日間（秒）
	})
}

// HandleRefreshToken はアクセストークンをリフレッシュします
// POST /token/refresh
func (h *TokenHandler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	// Authorization ヘッダーからトークンを取得
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		WriteError(w, http.StatusUnauthorized, "missing_token", "Authorization header is required")
		return
	}

	// Bearer トークンの形式を確認
	var tokenString string
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		WriteError(w, http.StatusUnauthorized, "invalid_token_format", "Authorization header must be in 'Bearer <token>' format")
		return
	}

	// トークンを検証
	claims, err := jwt.ValidateToken(tokenString, h.jwtSecret)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "invalid_token", "Token is invalid or expired")
		return
	}

	// 新しいJWTを生成（7日間有効）
	newAccessToken, err := jwt.GenerateToken(
		claims.UserID,
		claims.DiscordID,
		claims.Username,
		h.jwtSecret,
		7*24*time.Hour,
	)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "token_generation_failed", "Failed to generate new access token")
		return
	}

	// レスポンスを返す
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"access_token": newAccessToken,
		"token_type":   "Bearer",
		"expires_in":   7 * 24 * 3600, // 7日間（秒）
	})
}
