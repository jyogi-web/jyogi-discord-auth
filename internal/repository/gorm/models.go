package gorm

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
)

// User GORM model
type User struct {
	ID          string       `gorm:"primaryKey;type:varchar(36)"`
	DiscordID   string       `gorm:"uniqueIndex;type:varchar(255);not null"`
	Username    string       `gorm:"type:varchar(255);not null"`
	AvatarURL   string       `gorm:"type:varchar(512)"`
	CreatedAt   time.Time    `gorm:"autoCreateTime"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime"`
	LastLoginAt sql.NullTime `gorm:"type:datetime"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) ToDomain() *domain.User {
	var lastLoginAt *time.Time
	if u.LastLoginAt.Valid {
		lastLoginAt = &u.LastLoginAt.Time
	}
	return &domain.User{
		ID:          u.ID,
		DiscordID:   u.DiscordID,
		Username:    u.Username,
		AvatarURL:   u.AvatarURL,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
		LastLoginAt: lastLoginAt,
	}
}

func FromDomainUser(u *domain.User) *User {
	var lastLoginAt sql.NullTime
	if u.LastLoginAt != nil {
		lastLoginAt = sql.NullTime{Time: *u.LastLoginAt, Valid: true}
	}
	return &User{
		ID:          u.ID,
		DiscordID:   u.DiscordID,
		Username:    u.Username,
		AvatarURL:   u.AvatarURL,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
		LastLoginAt: lastLoginAt,
	}
}

// Session GORM model
type Session struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)"`
	UserID    string    `gorm:"index;type:varchar(36);not null"`
	Token     string    `gorm:"uniqueIndex;type:varchar(255);not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (Session) TableName() string {
	return "sessions"
}

func (s *Session) ToDomain() *domain.Session {
	return &domain.Session{
		ID:        s.ID,
		UserID:    s.UserID,
		Token:     s.Token,
		ExpiresAt: s.ExpiresAt,
		CreatedAt: s.CreatedAt,
	}
}

func FromDomainSession(s *domain.Session) *Session {
	return &Session{
		ID:        s.ID,
		UserID:    s.UserID,
		Token:     s.Token,
		ExpiresAt: s.ExpiresAt,
		CreatedAt: s.CreatedAt,
	}
}

// ClientApp GORM model
type ClientApp struct {
	ID           string    `gorm:"primaryKey;type:varchar(36)"`
	ClientID     string    `gorm:"uniqueIndex;type:varchar(255);not null"`
	ClientSecret string    `gorm:"type:varchar(255);not null"`
	Name         string    `gorm:"type:varchar(255);not null"`
	RedirectURIs string    `gorm:"type:text;not null"` // JSON string
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

func (ClientApp) TableName() string {
	return "client_apps"
}

func (c *ClientApp) ToDomain() (*domain.ClientApp, error) {
	var redirectURIs []string
	if err := json.Unmarshal([]byte(c.RedirectURIs), &redirectURIs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal redirect_uris: %w", err)
	}
	return &domain.ClientApp{
		ID:           c.ID,
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Name:         c.Name,
		RedirectURIs: redirectURIs,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}, nil
}

func FromDomainClientApp(c *domain.ClientApp) (*ClientApp, error) {
	redirectURIsJSON, err := json.Marshal(c.RedirectURIs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal redirect_uris: %w", err)
	}
	return &ClientApp{
		ID:           c.ID,
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Name:         c.Name,
		RedirectURIs: string(redirectURIsJSON),
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}, nil
}

// AuthCode GORM model
type AuthCode struct {
	ID          string    `gorm:"primaryKey;type:varchar(36)"`
	Code        string    `gorm:"uniqueIndex;type:varchar(255);not null"`
	ClientID    string    `gorm:"index;type:varchar(36);not null"` // Foreign key relationship handled logically
	UserID      string    `gorm:"index;type:varchar(36);not null"`
	RedirectURI string    `gorm:"type:text;not null"`
	ExpiresAt   time.Time `gorm:"not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	Used        bool      `gorm:"not null;default:false"`
}

func (AuthCode) TableName() string {
	return "auth_codes"
}

func (a *AuthCode) ToDomain() *domain.AuthCode {
	return &domain.AuthCode{
		ID:          a.ID,
		Code:        a.Code,
		ClientID:    a.ClientID,
		UserID:      a.UserID,
		RedirectURI: a.RedirectURI,
		ExpiresAt:   a.ExpiresAt,
		CreatedAt:   a.CreatedAt,
		Used:        a.Used,
	}
}

func FromDomainAuthCode(a *domain.AuthCode) *AuthCode {
	return &AuthCode{
		ID:          a.ID,
		Code:        a.Code,
		ClientID:    a.ClientID,
		UserID:      a.UserID,
		RedirectURI: a.RedirectURI,
		ExpiresAt:   a.ExpiresAt,
		CreatedAt:   a.CreatedAt,
		Used:        a.Used,
	}
}

// Token GORM model
type Token struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)"`
	Token     string    `gorm:"uniqueIndex;type:varchar(255);not null"`
	TokenType string    `gorm:"type:varchar(50);not null"`
	UserID    string    `gorm:"index;type:varchar(36);not null"`
	ClientID  string    `gorm:"index;type:varchar(36);not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	Revoked   bool      `gorm:"not null;default:false"`
}

func (Token) TableName() string {
	return "tokens"
}

func (t *Token) ToDomain() *domain.Token {
	return &domain.Token{
		ID:        t.ID,
		Token:     t.Token,
		TokenType: domain.TokenType(t.TokenType),
		UserID:    t.UserID,
		ClientID:  t.ClientID,
		ExpiresAt: t.ExpiresAt,
		CreatedAt: t.CreatedAt,
		Revoked:   t.Revoked,
	}
}

func FromDomainToken(t *domain.Token) *Token {
	return &Token{
		ID:        t.ID,
		Token:     t.Token,
		TokenType: string(t.TokenType),
		UserID:    t.UserID,
		ClientID:  t.ClientID,
		ExpiresAt: t.ExpiresAt,
		CreatedAt: t.CreatedAt,
		Revoked:   t.Revoked,
	}
}

// Profile GORM model
type Profile struct {
	ID               string    `gorm:"primaryKey;type:varchar(36)"`
	UserID           string    `gorm:"index;type:varchar(36);not null"`
	DiscordMessageID string    `gorm:"uniqueIndex;type:varchar(255);not null"`
	RealName         string    `gorm:"type:varchar(255)"`
	StudentID        string    `gorm:"type:varchar(255)"`
	Hobbies          string    `gorm:"type:text"`
	WhatToDo         string    `gorm:"type:text"`
	Comment          string    `gorm:"type:text"`
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime"`
}

func (Profile) TableName() string {
	return "profiles"
}

func (p *Profile) ToDomain() *domain.Profile {
	return &domain.Profile{
		ID:               p.ID,
		UserID:           p.UserID,
		DiscordMessageID: p.DiscordMessageID,
		RealName:         p.RealName,
		StudentID:        p.StudentID,
		Hobbies:          p.Hobbies,
		WhatToDo:         p.WhatToDo,
		Comment:          p.Comment,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}
}

func FromDomainProfile(p *domain.Profile) *Profile {
	return &Profile{
		ID:               p.ID,
		UserID:           p.UserID,
		DiscordMessageID: p.DiscordMessageID,
		RealName:         p.RealName,
		StudentID:        p.StudentID,
		Hobbies:          p.Hobbies,
		WhatToDo:         p.WhatToDo,
		Comment:          p.Comment,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}
}
