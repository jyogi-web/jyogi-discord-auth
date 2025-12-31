package domain

import (
	"fmt"
	"time"
)

// Session はユーザーのログインセッションを表します
type Session struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// Validate はセッションデータが有効かどうかを確認します
func (s *Session) Validate() error {
	if s.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if s.Token == "" {
		return fmt.Errorf("token is required")
	}
	if s.ExpiresAt.IsZero() {
		return fmt.Errorf("expires_at is required")
	}
	return nil
}

// IsExpired はセッションが期限切れかどうかを確認します
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
