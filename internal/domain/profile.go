package domain

import (
	"fmt"
	"time"
)

// Profile はじょぎメンバーのプロフィール情報を表します
type Profile struct {
	ID               string
	UserID           string
	DiscordMessageID string
	RealName         string
	StudentID        string
	Hobbies          string
	WhatToDo         string
	Comment          string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Validate はプロフィールデータが有効かどうかを確認します
func (p *Profile) Validate() error {
	if p.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if p.DiscordMessageID == "" {
		return fmt.Errorf("discord_message_id is required")
	}
	return nil
}
