package gorm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
	"github.com/jyogi-web/jyogi-discord-auth/internal/repository"
)

type tokenRepository struct {
	db *gorm.DB
}

// NewTokenRepository は新しいGORMトークンリポジトリを作成します
func NewTokenRepository(db *gorm.DB) repository.TokenRepository {
	return &tokenRepository{db: db}
}

// Create は新しいトークンをデータベースに挿入します
func (r *tokenRepository) Create(ctx context.Context, token *domain.Token) error {
	if err := token.Validate(); err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	t := FromDomainToken(token)
	if err := r.db.WithContext(ctx).Create(t).Error; err != nil {
		return fmt.Errorf("failed to create token: %w", err)
	}

	return nil
}

// GetByToken はトークン文字列でトークン情報を取得します
func (r *tokenRepository) GetByToken(ctx context.Context, token string) (*domain.Token, error) {
	var t Token
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("token not found")
		}
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	return t.ToDomain(), nil
}

// GetByUserID はユーザーIDに関連するすべてのトークンを取得します
func (r *tokenRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Token, error) {
	var tokens []Token
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&tokens).Error; err != nil {
		return nil, fmt.Errorf("failed to get tokens by user_id: %w", err)
	}

	domainTokens := make([]*domain.Token, len(tokens))
	for i, t := range tokens {
		domainTokens[i] = t.ToDomain()
	}

	return domainTokens, nil
}

// Revoke はトークンを無効化（取り消し）します
func (r *tokenRepository) Revoke(ctx context.Context, token string) error {
	result := r.db.WithContext(ctx).Model(&Token{}).Where("token = ?", token).Update("revoked", true)
	if result.Error != nil {
		return fmt.Errorf("failed to revoke token: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("token not found")
	}
	return nil
}

// DeleteExpired は期限切れのトークンを削除します
func (r *tokenRepository) DeleteExpired(ctx context.Context) error {
	result := r.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&Token{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete expired tokens: %w", result.Error)
	}
	return nil
}
