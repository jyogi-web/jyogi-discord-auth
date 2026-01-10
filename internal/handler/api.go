package handler

import (
	"log"
	"net/http"

	"github.com/jyogi-web/jyogi-discord-auth/internal/middleware"
	"github.com/jyogi-web/jyogi-discord-auth/internal/service"
)

// APIHandler はAPIハンドラーを表します
type APIHandler struct {
	authService *service.AuthService
}

// NewAPIHandler は新しいAPIハンドラーを作成します
func NewAPIHandler(authService *service.AuthService) *APIHandler {
	return &APIHandler{
		authService: authService,
	}
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

	// ユーザー情報とプロフィールを取得
	memberWithProfile, err := h.authService.GetUserWithProfile(r.Context(), claims.UserID)
	if err != nil {
		log.Printf("Failed to get user profile: %v", err)
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to get user info")
		return
	}

	// DTOに変換して返す
	dto := NewUserWithProfile(memberWithProfile.User, memberWithProfile.Profile)
	WriteJSON(w, http.StatusOK, dto)
}
