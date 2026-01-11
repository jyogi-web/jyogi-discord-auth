package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
	"github.com/jyogi-web/jyogi-discord-auth/internal/service"
)

// ClientHandler はクライアント管理ハンドラーを表します
type ClientHandler struct {
	clientService *service.ClientService
	authService   *service.AuthService
	templates     *template.Template
}

// NewClientHandler は新しいクライアント管理ハンドラーを作成します
func NewClientHandler(clientService *service.ClientService, authService *service.AuthService) *ClientHandler {
	// テンプレートをパース
	templates, err := template.ParseGlob("web/templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	return &ClientHandler{
		clientService: clientService,
		authService:   authService,
		templates:     templates,
	}
}

// HandleIndex はGET /を処理します
// ホーム画面を表示します
func (h *ClientHandler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	// セッション認証を試みる (任意)
	var user interface{}
	sessionCookie, err := r.Cookie("session_token")
	if err == nil {
		// セッションが存在する場合、ユーザー情報を取得
		u, err := h.authService.GetUserBySessionToken(r.Context(), sessionCookie.Value)
		if err == nil {
			user = map[string]interface{}{
				"Username":  u.Username,
				"AvatarURL": u.AvatarURL,
			}
		}
	}

	// テンプレートデータ
	data := map[string]interface{}{
		"User": user,
	}

	// テンプレートをレンダリング
	if err := h.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("Failed to render index template: %v", err)
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to render page")
		return
	}
}

// HandleRegisterForm はGET /clients/registerを処理します
// クライアント登録フォームを表示します
func (h *ClientHandler) HandleRegisterForm(w http.ResponseWriter, r *http.Request) {
	// GETメソッドのみ許可
	if r.Method != http.MethodGet {
		return // POSTは別ハンドラーで処理
	}

	// セッション認証
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		// 未認証の場合、ログイン画面にリダイレクト
		http.Redirect(w, r, "/auth/login?redirect_uri=/clients/register", http.StatusFound)
		return
	}

	// セッションを検証
	_, err = h.authService.GetUserBySessionToken(r.Context(), sessionCookie.Value)
	if err != nil {
		// セッション無効の場合もログイン画面にリダイレクト
		http.Redirect(w, r, "/auth/login?redirect_uri=/clients/register", http.StatusFound)
		return
	}

	// テンプレートデータ
	data := map[string]interface{}{
		"Error":        nil,
		"Name":         "",
		"RedirectURIs": "",
	}

	// テンプレートをレンダリング
	if err := h.templates.ExecuteTemplate(w, "register_client.html", data); err != nil {
		log.Printf("Failed to render template: %v", err)
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to render page")
		return
	}
}

// HandleRegisterSubmit はPOST /clients/registerを処理します
// クライアント登録処理を実行します
func (h *ClientHandler) HandleRegisterSubmit(w http.ResponseWriter, r *http.Request) {
	// POSTメソッドのみ許可
	if r.Method != http.MethodPost {
		return // GETは別ハンドラーで処理
	}

	// セッション認証
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		// 未認証の場合、ログイン画面にリダイレクト
		http.Redirect(w, r, "/auth/login?redirect_uri=/clients/register", http.StatusFound)
		return
	}

	// セッションを検証
	_, err = h.authService.GetUserBySessionToken(r.Context(), sessionCookie.Value)
	if err != nil {
		// セッション無効の場合もログイン画面にリダイレクト
		http.Redirect(w, r, "/auth/login?redirect_uri=/clients/register", http.StatusFound)
		return
	}

	// フォームパラメータを解析
	if err := r.ParseForm(); err != nil {
		h.renderFormWithError(w, "フォームの解析に失敗しました", "", "")
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	redirectURIsRaw := strings.TrimSpace(r.FormValue("redirect_uris"))

	// バリデーション: 必須フィールド
	if name == "" || redirectURIsRaw == "" {
		h.renderFormWithError(w, "クライアント名とリダイレクトURIは必須です", name, redirectURIsRaw)
		return
	}

	// バリデーション: 名前の長さ
	if len(name) > 255 {
		h.renderFormWithError(w, "クライアント名は255文字以内で入力してください", name, redirectURIsRaw)
		return
	}

	// リダイレクトURIをパース (改行区切り)
	lines := strings.Split(redirectURIsRaw, "\n")
	redirectURIs := make([]string, 0)
	for _, line := range lines {
		uri := strings.TrimSpace(line)
		if uri == "" {
			continue
		}
		redirectURIs = append(redirectURIs, uri)
	}

	// バリデーション: リダイレクトURIが最低1個必要
	if len(redirectURIs) == 0 {
		h.renderFormWithError(w, "最低1個のリダイレクトURIが必要です", name, redirectURIsRaw)
		return
	}

	// バリデーション: URL形式のチェック
	for _, uri := range redirectURIs {
		parsedURL, err := url.Parse(uri)
		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			h.renderFormWithError(w, fmt.Sprintf("不正なURL形式です: %s", uri), name, redirectURIsRaw)
			return
		}

		// HTTPSのみ許可 (開発環境では http://localhost も許可)
		if parsedURL.Scheme != "https" {
			if !(parsedURL.Scheme == "http" && strings.HasPrefix(parsedURL.Host, "localhost")) {
				h.renderFormWithError(w, fmt.Sprintf("HTTPSを使用してください (開発環境ではhttp://localhostのみ許可): %s", uri), name, redirectURIsRaw)
				return
			}
		}
	}

	// Client IDとSecretを自動生成
	clientID := uuid.New().String()
	clientSecret, err := generateClientSecret()
	if err != nil {
		log.Printf("Failed to generate client secret: %v", err)
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to generate credentials")
		return
	}

	// ClientServiceでクライアントを登録
	client, err := h.clientService.RegisterClient(r.Context(), clientID, clientSecret, name, redirectURIs)
	if err != nil {
		log.Printf("Failed to register client: %v", err)
		// エラーメッセージをユーザーに表示
		errorMsg := "クライアントの登録に失敗しました"
		if strings.Contains(err.Error(), "already exists") {
			errorMsg = "Client IDが既に存在します。もう一度お試しください。"
		}
		h.renderFormWithError(w, errorMsg, name, redirectURIsRaw)
		return
	}

	// 登録成功画面を表示 (平文シークレットを含む)
	data := map[string]interface{}{
		"Name":         client.Name,
		"ClientID":     client.ClientID,
		"ClientSecret": clientSecret, // 平文 (ここでのみ表示)
		"RedirectURIs": client.RedirectURIs,
	}

	if err := h.templates.ExecuteTemplate(w, "register_success.html", data); err != nil {
		log.Printf("Failed to render success template: %v", err)
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to render success page")
		return
	}
}

// renderFormWithError はエラーメッセージ付きでフォームを再表示します
func (h *ClientHandler) renderFormWithError(w http.ResponseWriter, errorMsg, name, redirectURIs string) {
	data := map[string]interface{}{
		"Error":        errorMsg,
		"Name":         name,
		"RedirectURIs": redirectURIs,
	}

	w.WriteHeader(http.StatusBadRequest)
	if err := h.templates.ExecuteTemplate(w, "register_client.html", data); err != nil {
		log.Printf("Failed to render error template: %v", err)
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to render error page")
	}
}

// generateClientSecret は32バイトのランダムなClient Secretを生成します
func generateClientSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate secret: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// HandleListClients はGET /clientsを処理します
// 登録済みクライアント一覧を表示します
func (h *ClientHandler) HandleListClients(w http.ResponseWriter, r *http.Request) {
	// セッション認証
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		// 未認証の場合、ログイン画面にリダイレクト
		http.Redirect(w, r, "/auth/login?redirect_uri=/clients", http.StatusFound)
		return
	}

	// セッションを検証
	_, err = h.authService.GetUserBySessionToken(r.Context(), sessionCookie.Value)
	if err != nil {
		// セッション無効の場合もログイン画面にリダイレクト
		http.Redirect(w, r, "/auth/login?redirect_uri=/clients", http.StatusFound)
		return
	}

	// 全クライアントを取得
	clients, err := h.clientService.GetAllClients(r.Context())
	if err != nil {
		log.Printf("Failed to get clients: %v", err)
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to get clients")
		return
	}

	// テンプレートデータ
	data := map[string]interface{}{
		"Clients": clients,
	}

	// テンプレートをレンダリング
	if err := h.templates.ExecuteTemplate(w, "clients_list.html", data); err != nil {
		log.Printf("Failed to render clients list template: %v", err)
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to render page")
		return
	}
}

// HandleEditClientForm はGET /clients/:id/editを処理します
// クライアント編集フォームを表示します
func (h *ClientHandler) HandleEditClientForm(w http.ResponseWriter, r *http.Request) {
	// セッション認証
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		http.Redirect(w, r, "/auth/login?redirect_uri="+r.URL.Path, http.StatusFound)
		return
	}

	_, err = h.authService.GetUserBySessionToken(r.Context(), sessionCookie.Value)
	if err != nil {
		http.Redirect(w, r, "/auth/login?redirect_uri="+r.URL.Path, http.StatusFound)
		return
	}

	// URLからクライアントIDを取得 (パスは /clients/:id/edit)
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid client ID")
		return
	}
	clientID := pathParts[1]

	// クライアント情報を取得
	client, err := h.clientService.GetClientByID(r.Context(), clientID)
	if err != nil {
		log.Printf("Failed to get client: %v", err)
		WriteError(w, http.StatusNotFound, "not_found", "Client not found")
		return
	}

	// RedirectURIsを改行区切りのテキストに変換
	redirectURIsText := strings.Join(client.RedirectURIs, "\n")

	// テンプレートデータ
	data := map[string]interface{}{
		"Client":           client,
		"RedirectURIsText": redirectURIsText,
		"Error":            nil,
	}

	// テンプレートをレンダリング
	if err := h.templates.ExecuteTemplate(w, "edit_client.html", data); err != nil {
		log.Printf("Failed to render edit template: %v", err)
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to render page")
		return
	}
}

// HandleUpdateClient はPOST /clients/:idを処理します
// クライアント情報を更新します
func (h *ClientHandler) HandleUpdateClient(w http.ResponseWriter, r *http.Request) {
	// セッション認証
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		http.Redirect(w, r, "/auth/login", http.StatusFound)
		return
	}

	_, err = h.authService.GetUserBySessionToken(r.Context(), sessionCookie.Value)
	if err != nil {
		http.Redirect(w, r, "/auth/login", http.StatusFound)
		return
	}

	// URLからクライアントIDを取得
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid client ID")
		return
	}
	clientID := pathParts[1]

	// 既存のクライアント情報を取得
	client, err := h.clientService.GetClientByID(r.Context(), clientID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "not_found", "Client not found")
		return
	}

	// フォームパラメータを解析
	if err := r.ParseForm(); err != nil {
		h.renderEditFormWithError(w, client, "フォームの解析に失敗しました", "")
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	redirectURIsRaw := strings.TrimSpace(r.FormValue("redirect_uris"))

	// バリデーション
	if name == "" || redirectURIsRaw == "" {
		h.renderEditFormWithError(w, client, "クライアント名とリダイレクトURIは必須です", redirectURIsRaw)
		return
	}

	if len(name) > 255 {
		h.renderEditFormWithError(w, client, "クライアント名は255文字以内で入力してください", redirectURIsRaw)
		return
	}

	// リダイレクトURIをパース
	lines := strings.Split(redirectURIsRaw, "\n")
	redirectURIs := make([]string, 0)
	for _, line := range lines {
		uri := strings.TrimSpace(line)
		if uri == "" {
			continue
		}
		redirectURIs = append(redirectURIs, uri)
	}

	if len(redirectURIs) == 0 {
		h.renderEditFormWithError(w, client, "最低1個のリダイレクトURIが必要です", redirectURIsRaw)
		return
	}

	// URL形式のチェック
	for _, uri := range redirectURIs {
		parsedURL, err := url.Parse(uri)
		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			h.renderEditFormWithError(w, client, fmt.Sprintf("不正なURL形式です: %s", uri), redirectURIsRaw)
			return
		}

		if parsedURL.Scheme != "https" {
			if !(parsedURL.Scheme == "http" && strings.HasPrefix(parsedURL.Host, "localhost")) {
				h.renderEditFormWithError(w, client, fmt.Sprintf("HTTPSを使用してください: %s", uri), redirectURIsRaw)
				return
			}
		}
	}

	// ClientServiceで更新 (Secretは変更しない)
	_, err = h.clientService.UpdateClient(r.Context(), client.ClientID, "", name, redirectURIs)
	if err != nil {
		log.Printf("Failed to update client: %v", err)
		h.renderEditFormWithError(w, client, "クライアントの更新に失敗しました", redirectURIsRaw)
		return
	}

	// 一覧画面にリダイレクト
	http.Redirect(w, r, "/clients", http.StatusFound)
}

// HandleDeleteClient はDELETE /clients/:idを処理します
// クライアントを削除します
func (h *ClientHandler) HandleDeleteClient(w http.ResponseWriter, r *http.Request) {
	// セッション認証
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"message": "Unauthorized",
		})
		return
	}

	_, err = h.authService.GetUserBySessionToken(r.Context(), sessionCookie.Value)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"message": "Invalid session",
		})
		return
	}

	// URLからクライアントIDを取得
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid client ID",
		})
		return
	}
	clientID := pathParts[1]

	// ClientServiceで削除
	err = h.clientService.DeleteClient(r.Context(), clientID)
	if err != nil {
		log.Printf("Failed to delete client: %v", err)
		WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Failed to delete client",
		})
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Client deleted successfully",
	})
}

// renderEditFormWithError はエラーメッセージ付きで編集フォームを再表示します
func (h *ClientHandler) renderEditFormWithError(w http.ResponseWriter, client *domain.ClientApp, errorMsg, redirectURIsText string) {
	if redirectURIsText == "" {
		redirectURIsText = strings.Join(client.RedirectURIs, "\n")
	}

	data := map[string]interface{}{
		"Client":           client,
		"RedirectURIsText": redirectURIsText,
		"Error":            errorMsg,
	}

	w.WriteHeader(http.StatusBadRequest)
	if err := h.templates.ExecuteTemplate(w, "edit_client.html", data); err != nil {
		log.Printf("Failed to render error template: %v", err)
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to render error page")
	}
}
