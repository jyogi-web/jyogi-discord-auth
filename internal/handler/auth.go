package handler

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
	"github.com/jyogi-web/jyogi-discord-auth/internal/service"
)

// AuthHandler は認証ハンドラーを表します
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler は新しい認証ハンドラーを作成します
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// HandleLogin はログインリクエストを処理します
// Discord OAuth2認証ページにリダイレクトします
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// クライアントアプリから渡されたredirect_uriを取得
	redirectURI := r.URL.Query().Get("redirect_uri")
	if redirectURI == "" {
		// デフォルトのリダイレクト先（デモアプリ）
		redirectURI = "http://localhost:3000/auth/callback"
	}

	// CSRF攻撃を防ぐためのstateを生成
	state, err := h.authService.GenerateState()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to generate state")
		return
	}

	// stateをセッションに保存（Cookieを使用）
	SetSecureCookie(w, r, CookieOptions{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600, // 10分間有効
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// redirect_uriもCookieに保存（コールバック時に使用）
	SetSecureCookie(w, r, CookieOptions{
		Name:     "redirect_uri",
		Value:    redirectURI,
		Path:     "/",
		MaxAge:   600, // 10分間有効
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// Discord認証URLを生成してリダイレクト
	authURL := h.authService.GetAuthURL(state)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// HandleCallback はDiscord OAuth2コールバックを処理します
func (h *AuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	// クエリパラメータを取得
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		WriteError(w, http.StatusBadRequest, "missing_code", "Authorization code is required")
		return
	}

	if state == "" {
		WriteError(w, http.StatusBadRequest, "missing_state", "State parameter is required")
		return
	}

	// Cookieからstateを取得して検証
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_state", "State cookie not found")
		return
	}

	if stateCookie.Value != state {
		WriteError(w, http.StatusBadRequest, "state_mismatch", "State parameter does not match")
		return
	}

	// state cookieを削除
	DeleteCookie(w, r, "oauth_state", "/")

	// コールバックを処理してセッションを作成
	sessionToken, err := h.authService.HandleCallback(r.Context(), code)
	if err != nil {
		// じょぎメンバーでない場合
		if errors.Is(err, domain.ErrNotGuildMember) {
			WriteError(w, http.StatusForbidden, "not_guild_member", "You are not a member of the じょぎ server")
			return
		}

		// エラー内容をログに出力
		log.Printf("Callback error: %v", err)
		WriteError(w, http.StatusInternalServerError, "callback_failed", "Failed to process authentication callback")
		return
	}

	// セッショントークンをCookieに保存
	SetSecureCookie(w, r, CookieOptions{
		Name:     "session_token",
		Value:    sessionToken,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 7日間
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// Cookieから保存されたredirect_uriを取得
	redirectCookie, err := r.Cookie("redirect_uri")
	redirectURL := "http://localhost:3000/auth/callback" // デフォルト
	if err == nil && redirectCookie.Value != "" {
		redirectURL = redirectCookie.Value
	}

	// redirect_uri Cookieを削除（使用済み）
	DeleteCookie(w, r, "redirect_uri", "/")

	// クライアントアプリにリダイレクト
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// HandleLogout はログアウトリクエストを処理します
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// セッショントークンを取得
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "no_session", "No active session")
		return
	}

	// セッションを削除
	if err := h.authService.Logout(r.Context(), sessionCookie.Value); err != nil {
		WriteError(w, http.StatusInternalServerError, "logout_failed", "Failed to logout")
		return
	}

	// セッションCookieを削除
	DeleteCookie(w, r, "session_token", "/")

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Logout successful",
	})
}

// HandleMe は現在のユーザー情報を取得します
func (h *AuthHandler) HandleMe(w http.ResponseWriter, r *http.Request) {
	// セッショントークンを取得
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "No active session")
		return
	}

	// ユーザー情報を取得
	user, err := h.authService.GetUserBySessionToken(r.Context(), sessionCookie.Value)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "invalid_session", "Session is invalid or expired")
		return
	}

	// ユーザー情報を返す
	response := map[string]interface{}{
		"id":         user.ID,
		"discord_id": user.DiscordID,
		"username":   user.Username,
		"avatar_url": user.AvatarURL,
	}

	if user.LastLoginAt != nil {
		response["last_login_at"] = user.LastLoginAt.Format(time.RFC3339)
	}

	WriteJSON(w, http.StatusOK, response)
}

// HandleMembers はじょぎメンバー一覧を返します
// GET /api/members
func (h *AuthHandler) HandleMembers(w http.ResponseWriter, r *http.Request) {
	// セッショントークンを取得
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "No active session")
		return
	}

	// セッションを検証
	_, err = h.authService.GetUserBySessionToken(r.Context(), sessionCookie.Value)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Invalid or expired session")
		return
	}

	// メンバー一覧をプロフィール情報付きで取得
	membersWithProfiles, err := h.authService.GetAllMembersWithProfiles(r.Context())
	if err != nil {
		log.Printf("Failed to get members: %v", err)
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to get members")
		return
	}

	// DTOに変換
	membersList := make([]*UserWithProfile, len(membersWithProfiles))
	for i, memberWithProfile := range membersWithProfiles {
		membersList[i] = NewUserWithProfile(memberWithProfile.User, memberWithProfile.Profile)
	}

	// メンバー一覧を返す
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"members": membersList,
		"total":   len(membersList),
	})
}
