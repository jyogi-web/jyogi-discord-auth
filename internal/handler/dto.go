package handler

import (
	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
)

// UserWithProfile はユーザー情報とプロフィール情報を結合したDTO
type UserWithProfile struct {
	// ユーザー基本情報
	ID          string  `json:"id"`
	DiscordID   string  `json:"discord_id"`
	Username    string  `json:"username"`
	AvatarURL   string  `json:"avatar_url"`
	LastLoginAt *string `json:"last_login_at,omitempty"`

	// プロフィール情報（オプション）
	Profile *ProfileData `json:"profile,omitempty"`
}

// ProfileData はプロフィール情報のDTO
type ProfileData struct {
	RealName  *string `json:"real_name,omitempty"`
	StudentID *string `json:"student_id,omitempty"`
	Hobbies   *string `json:"hobbies,omitempty"`
	WhatToDo  *string `json:"what_to_do,omitempty"`
	Comment   *string `json:"comment,omitempty"`
}

// NewUserWithProfile はドメインモデルからDTOを作成します
func NewUserWithProfile(user *domain.User, profile *domain.Profile) *UserWithProfile {
	dto := &UserWithProfile{
		ID:        user.ID,
		DiscordID: user.DiscordID,
		Username:  user.Username,
		AvatarURL: user.AvatarURL,
	}

	// LastLoginAtの変換
	if user.LastLoginAt != nil {
		lastLogin := user.LastLoginAt.Format("2006-01-02T15:04:05Z07:00")
		dto.LastLoginAt = &lastLogin
	}

	// プロフィール情報が存在する場合は追加
	if profile != nil {
		dto.Profile = &ProfileData{
			RealName:  stringToPtr(profile.RealName),
			StudentID: stringToPtr(profile.StudentID),
			Hobbies:   stringToPtr(profile.Hobbies),
			WhatToDo:  stringToPtr(profile.WhatToDo),
			Comment:   stringToPtr(profile.Comment),
		}
	}

	return dto
}

// stringToPtr は空文字列でなければポインタを返す
func stringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
