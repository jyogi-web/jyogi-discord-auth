package handler

import (
	"log"
	"net/http"
	"net/url"
	"strconv"

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
		// ログイン後にこのauthorizeリクエストに戻るように、現在のURLをredirect_uriとして保存
		SetSecureCookie(w, r, CookieOptions{
			Name:     "redirect_uri",
			Value:    r.URL.String(),
			Path:     "/",
			MaxAge:   600, // 10分間有効
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, "/auth/login", http.StatusFound)
		return
	}

	// セッションからユーザーを取得
	user, err := h.authService.GetUserBySessionToken(r.Context(), sessionCookie.Value)
	if err != nil {
		// セッションが無効な場合もログインにリダイレクト
		SetSecureCookie(w, r, CookieOptions{
			Name:     "redirect_uri",
			Value:    r.URL.String(),
			Path:     "/",
			MaxAge:   600, // 10分間有効
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, "/auth/login", http.StatusFound)
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

	// Authorization ヘッダーからトークンを取得
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"error":   "invalid_token",
			"message": "Authorization header is required",
		})
		return
	}

	// Bearer トークンの形式を確認
	var accessToken string
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		accessToken = authHeader[7:]
	} else {
		WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"error":   "invalid_token",
			"message": "Authorization header must be in 'Bearer <token>' format",
		})
		return
	}

	// アクセストークンからユーザー情報を取得
	user, err := h.oauth2Service.GetUserByAccessToken(r.Context(), accessToken)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"error":   "invalid_token",
			"message": "Token is invalid or expired",
		})
		return
	}

	// ユーザー情報とプロフィールを取得
	memberWithProfile, err := h.authService.GetUserWithProfile(r.Context(), user.ID)
	if err != nil {
		log.Printf("Failed to get user profile: %v", err)
		WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"error":   "internal_error",
			"message": "Failed to get user info",
		})
		return
	}

	// DTOに変換して返す（/api/userと同じ形式）
	dto := NewUserWithProfile(memberWithProfile.User, memberWithProfile.Profile)
	WriteJSON(w, http.StatusOK, dto)
}

// HandleUserByID はGET /oauth/user/{id}を処理します
// アクセストークンで認証し、指定されたIDのユーザー情報を返します
func (h *OAuth2Handler) HandleUserByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authorization ヘッダーからトークンを取得
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"error":   "invalid_token",
			"message": "Authorization header is required",
		})
		return
	}

	// Bearer トークンの形式を確認
	var accessToken string
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		accessToken = authHeader[7:]
	} else {
		WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"error":   "invalid_token",
			"message": "Authorization header must be in 'Bearer <token>' format",
		})
		return
	}

	// アクセストークンを検証（トークンの有効性を確認）
	_, err := h.oauth2Service.GetUserByAccessToken(r.Context(), accessToken)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"error":   "invalid_token",
			"message": "Token is invalid or expired",
		})
		return
	}

	// URLパラメータからIDを取得
	userID := r.PathValue("id")
	if userID == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "invalid_user_id",
			"message": "User ID is required",
		})
		return
	}

	// 指定されたIDのユーザー情報とプロフィールを取得
	memberWithProfile, err := h.authService.GetUserWithProfile(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get user profile: %v", err)
		WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"error":   "user_not_found",
			"message": "User not found",
		})
		return
	}

	// DTOに変換して返す
	dto := NewUserWithProfile(memberWithProfile.User, memberWithProfile.Profile)
	WriteJSON(w, http.StatusOK, dto)
}

// HandleMembers はGET /oauth/membersを処理します
// アクセストークンで認証し、じょぎメンバー一覧をプロフィール情報付きで返します
func (h *OAuth2Handler) HandleMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authorization ヘッダーからトークンを取得
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"error":   "invalid_token",
			"message": "Authorization header is required",
		})
		return
	}

	// Bearer トークンの形式を確認
	var accessToken string
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		accessToken = authHeader[7:]
	} else {
		WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"error":   "invalid_token",
			"message": "Authorization header must be in 'Bearer <token>' format",
		})
		return
	}

	// アクセストークンを検証（トークンの有効性を確認）
	_, err := h.oauth2Service.GetUserByAccessToken(r.Context(), accessToken)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"error":   "invalid_token",
			"message": "Token is invalid or expired",
		})
		return
	}

	// ページネーションパラメータの取得
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			if parsedLimit > 0 && parsedLimit <= 100 {
				limit = parsedLimit
			}
		}
	}

	offset := 0
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil {
			if parsedOffset >= 0 {
				offset = parsedOffset
			}
		}
	}

	// メンバー一覧をプロフィール情報付きで取得
	membersWithProfiles, err := h.authService.GetMembersWithProfiles(r.Context(), limit, offset)
	if err != nil {
		log.Printf("Failed to get members: %v", err)
		WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"error":   "internal_error",
			"message": "Failed to get members",
		})
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
		"limit":   limit,
		"offset":  offset,
		"count":   len(membersList),
	})
}
