package domain

import (
	"fmt"
	"time"
)

// AuthCode はOAuth2認可コードを表します
type AuthCode struct {
	ID          string
	Code        string
	ClientID    string
	UserID      string
	RedirectURI string
	ExpiresAt   time.Time
	CreatedAt   time.Time
	Used        bool
}

// Validate は認可コードのデータが有効かどうかを確認します
func (a *AuthCode) Validate() error {
	if a.Code == "" {
		return fmt.Errorf("code is required")
	}
	if a.ClientID == "" {
		return fmt.Errorf("client_id is required")
	}
	if a.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if a.RedirectURI == "" {
		return fmt.Errorf("redirect_uri is required")
	}
	if a.ExpiresAt.IsZero() {
		return fmt.Errorf("expires_at is required")
	}
	return nil
}

// IsExpired は認可コードが期限切れかどうかを確認します
func (a *AuthCode) IsExpired() bool {
	return time.Now().After(a.ExpiresAt)
}

// IsValid は認可コードが有効（期限切れでなく未使用）かどうかを確認します
func (a *AuthCode) IsValid() bool {
	return !a.IsExpired() && !a.Used
}
