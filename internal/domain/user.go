package domain

import (
	"fmt"
	"time"
)

// User はじょぎメンバーのユーザーを表します
type User struct {
	ID          string
	DiscordID   string
	Username    string
	AvatarURL   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	LastLoginAt *time.Time
}

// Validate はユーザーデータが有効かどうかを確認します
func (u *User) Validate() error {
	if u.DiscordID == "" {
		return fmt.Errorf("discord_id is required")
	}
	if u.Username == "" {
		return fmt.Errorf("username is required")
	}
	return nil
}
