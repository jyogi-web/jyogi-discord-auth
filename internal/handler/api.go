package handler

import (
	"net/http"

	"github.com/jyogi-web/jyogi-discord-auth/internal/middleware"
)

// APIHandler はAPIハンドラーを表します
type APIHandler struct{}

// NewAPIHandler は新しいAPIハンドラーを作成します
func NewAPIHandler() *APIHandler {
	return &APIHandler{}
}

// HandleVerify はJWTトークンの検証を行います
// GET /api/verify
func (h *APIHandler) HandleVerify(w http.ResponseWriter, r *http.Request) {
	// ミドルウェアで既にJWTが検証されているため、クレームを取得するだけ
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Failed to get user claims")
		return
	}

	// 検証成功レスポンス
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"valid":      true,
		"user_id":    claims.UserID,
		"discord_id": claims.DiscordID,
		"username":   claims.Username,
	})
}

// HandleUser はJWT認証されたユーザー情報を返します
// GET /api/user
func (h *APIHandler) HandleUser(w http.ResponseWriter, r *http.Request) {
	// ミドルウェアで既にJWTが検証されているため、クレームを取得するだけ
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Failed to get user claims")
		return
	}

	// ユーザー情報を返す
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":         claims.UserID,
		"discord_id": claims.DiscordID,
		"username":   claims.Username,
	})
}
