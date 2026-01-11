package handler

import "net/http"

// CookieOptions はCookieの設定オプションです
type CookieOptions struct {
	Name     string
	Value    string
	Path     string
	MaxAge   int
	HttpOnly bool
	SameSite http.SameSite
}

// SetSecureCookie はセキュアなCookieを設定します
// HTTPS接続の場合はSecureフラグを設定します
func SetSecureCookie(w http.ResponseWriter, r *http.Request, opts CookieOptions) {
	// Cloud RunやリバースプロキシではX-Forwarded-Protoヘッダーをチェック
	isHTTPS := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"

	cookie := &http.Cookie{
		Name:     opts.Name,
		Value:    opts.Value,
		Path:     opts.Path,
		MaxAge:   opts.MaxAge,
		HttpOnly: opts.HttpOnly,
		Secure:   isHTTPS, // HTTPS接続の場合はSecureフラグを設定
		SameSite: opts.SameSite,
	}

	http.SetCookie(w, cookie)
}

// DeleteCookie はCookieを削除します
func DeleteCookie(w http.ResponseWriter, r *http.Request, name, path string) {
	SetSecureCookie(w, r, CookieOptions{
		Name:     name,
		Value:    "",
		Path:     path,
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
