package repository

import (
	"context"

	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
)

// UserRepository はユーザーデータアクセスのインターフェースを定義します
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByDiscordID(ctx context.Context, discordID string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
	GetAll(ctx context.Context, limit, offset int) ([]*domain.User, error)
}

// SessionRepository はセッションデータアクセスのインターフェースを定義します
type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	GetByID(ctx context.Context, id string) (*domain.Session, error)
	GetByToken(ctx context.Context, token string) (*domain.Session, error)
	GetByUserID(ctx context.Context, userID string) ([]*domain.Session, error)
	Delete(ctx context.Context, id string) error
	DeleteByToken(ctx context.Context, token string) error
	DeleteExpired(ctx context.Context) error
}

// ClientRepository はクライアントアプリデータアクセスのインターフェースを定義します
type ClientRepository interface {
	Create(ctx context.Context, client *domain.ClientApp) error
	GetByID(ctx context.Context, id string) (*domain.ClientApp, error)
	GetByClientID(ctx context.Context, clientID string) (*domain.ClientApp, error)
	GetAll(ctx context.Context) ([]*domain.ClientApp, error)
	Update(ctx context.Context, client *domain.ClientApp) error
	Delete(ctx context.Context, id string) error
	ValidateRedirectURI(ctx context.Context, clientID, redirectURI string) (bool, error)
}

// AuthCodeRepository は認可コードデータアクセスのインターフェースを定義します
type AuthCodeRepository interface {
	Create(ctx context.Context, authCode *domain.AuthCode) error
	GetByCode(ctx context.Context, code string) (*domain.AuthCode, error)
	MarkAsUsed(ctx context.Context, code string) error
	DeleteExpired(ctx context.Context) error
}

// TokenRepository はトークンデータアクセスのインターフェースを定義します
type TokenRepository interface {
	Create(ctx context.Context, token *domain.Token) error
	GetByToken(ctx context.Context, token string) (*domain.Token, error)
	GetByUserID(ctx context.Context, userID string) ([]*domain.Token, error)
	Revoke(ctx context.Context, token string) error
	DeleteExpired(ctx context.Context) error
}

// ProfileRepository はプロフィールデータアクセスのインターフェースを定義します
type ProfileRepository interface {
	Create(ctx context.Context, profile *domain.Profile) error
	GetByID(ctx context.Context, id string) (*domain.Profile, error)
	GetByUserID(ctx context.Context, userID string) (*domain.Profile, error)
	GetByMessageID(ctx context.Context, messageID string) (*domain.Profile, error)
	GetAll(ctx context.Context) ([]*domain.Profile, error)
	Update(ctx context.Context, profile *domain.Profile) error
	Upsert(ctx context.Context, profile *domain.Profile) error
	Delete(ctx context.Context, id string) error
}
