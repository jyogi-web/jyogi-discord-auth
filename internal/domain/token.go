package domain

import (
	"fmt"
	"time"
)

// TokenType はトークンの種類を表します
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// Token はアクセストークンまたはリフレッシュトークンを表します
type Token struct {
	ID        string
	Token     string
	TokenType TokenType
	UserID    string
	ClientID  string
	ExpiresAt time.Time
	CreatedAt time.Time
	Revoked   bool
}

// Validate はトークンデータが有効かどうかを確認します
func (t *Token) Validate() error {
	if t.Token == "" {
		return fmt.Errorf("token is required")
	}
	if t.TokenType != TokenTypeAccess && t.TokenType != TokenTypeRefresh {
		return fmt.Errorf("invalid token_type: must be 'access' or 'refresh'")
	}
	if t.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if t.ClientID == "" {
		return fmt.Errorf("client_id is required")
	}
	if t.ExpiresAt.IsZero() {
		return fmt.Errorf("expires_at is required")
	}
	return nil
}

// IsExpired はトークンが期限切れかどうかを確認します
func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsValid はトークンが有効（期限切れでなく取り消されていない）かどうかを確認します
func (t *Token) IsValid() bool {
	return !t.IsExpired() && !t.Revoked
}
