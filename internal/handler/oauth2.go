package handler

import (
	"net/http"
	"net/url"

	"github.com/jyogi-web/jyogi-discord-auth/internal/service"
)

// OAuth2Handler はOAuth2エンドポイントのハンドラーです
type OAuth2Handler struct {
	oauth2Service *service.OAuth2Service
	authService   *service.AuthService
}

// NewOAuth2Handler は新しいOAuth2ハンドラーを作成します
func NewOAuth2Handler(oauth2Service *service.OAuth2Service, authService *service.AuthService) *OAuth2Handler {
	return &OAuth2Handler{
		oauth2Service: oauth2Service,
		authService:   authService,
	}
}

// HandleAuthorize はGET /oauth/authorizeを処理します
// クライアントアプリからの認可リクエストを受け付けます
func (h *OAuth2Handler) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// クエリパラメータを取得
	query := r.URL.Query()
	clientID := query.Get("client_id")
	redirectURI := query.Get("redirect_uri")
	responseType := query.Get("response_type")
	state := query.Get("state")

	// 必須パラメータのチェック
	if clientID == "" || redirectURI == "" || responseType == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":             "invalid_request",
			"error_description": "missing required parameters",
		})
		return
	}

	// ユーザーがログインしているかチェック（セッションから取得）
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		// ユーザーが未ログインの場合、Discordログインにリダイレクト
		// ログイン後にこのauthorizeリクエストに戻るようにstate情報を保存する必要がある
		WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"error":             "login_required",
			"error_description": "user must be logged in",
		})
		return
	}

	// セッションからユーザーを取得
	user, err := h.authService.GetUserBySessionToken(r.Context(), sessionCookie.Value)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"error":             "invalid_session",
			"error_description": "session is invalid or expired",
		})
		return
	}

	// 認可リクエストを処理
	authReq := &service.AuthorizeRequest{
		ClientID:     clientID,
		RedirectURI:  redirectURI,
		ResponseType: responseType,
		State:        state,
		UserID:       user.ID,
	}

	authResp, err := h.oauth2Service.Authorize(r.Context(), authReq)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":             "invalid_request",
			"error_description": err.Error(),
		})
		return
	}

	// 認可コードとstateをクエリパラメータに含めてredirect_uriにリダイレクト
	redirectURL, err := url.Parse(authResp.RedirectURI)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"error":             "server_error",
			"error_description": "failed to parse redirect_uri",
		})
		return
	}

	params := redirectURL.Query()
	params.Add("code", authResp.Code)
	if authResp.State != "" {
		params.Add("state", authResp.State)
	}
	redirectURL.RawQuery = params.Encode()

	http.Redirect(w, r, redirectURL.String(), http.StatusFound)
}

// HandleToken はPOST /oauth/tokenを処理します
// 認可コードをアクセストークンに交換します
func (h *OAuth2Handler) HandleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Content-Typeをチェック
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/x-www-form-urlencoded" {
		WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":             "invalid_request",
			"error_description": "Content-Type must be application/x-www-form-urlencoded",
		})
		return
	}

	// フォームパラメータを解析
	if err := r.ParseForm(); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":             "invalid_request",
			"error_description": "failed to parse form",
		})
		return
	}

	grantType := r.FormValue("grant_type")
	code := r.FormValue("code")
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")
	redirectURI := r.FormValue("redirect_uri")

	// 必須パラメータのチェック
	if grantType == "" || code == "" || clientID == "" || clientSecret == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":             "invalid_request",
			"error_description": "missing required parameters",
		})
		return
	}

	// トークンリクエストを処理
	tokenReq := &service.TokenRequest{
		GrantType:    grantType,
		Code:         code,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURI,
	}

	tokenResp, err := h.oauth2Service.ExchangeToken(r.Context(), tokenReq)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":             "invalid_grant",
			"error_description": err.Error(),
		})
		return
	}

	// レスポンスを返す
	WriteJSON(w, http.StatusOK, tokenResp)
}

// HandleVerifyToken はGET /oauth/verify を処理します
// アクセストークンの検証エンドポイント
func (h *OAuth2Handler) HandleVerifyToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 一時的にエラーを返す
	WriteJSON(w, http.StatusNotImplemented, map[string]interface{}{
		"error":             "not_implemented",
		"error_description": "token verification not yet implemented",
	})
}

// HandleUserInfo はGET /oauth/userinfoを処理します
// アクセストークンに紐づくユーザー情報を返します
func (h *OAuth2Handler) HandleUserInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 一時的にエラーを返す
	WriteJSON(w, http.StatusNotImplemented, map[string]interface{}{
		"error":             "not_implemented",
		"error_description": "userinfo endpoint not yet implemented",
	})
}
