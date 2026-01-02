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

type authCodeRepository struct {
	db *gorm.DB
}

// NewAuthCodeRepository は新しいGORM認可コードリポジトリを作成します
func NewAuthCodeRepository(db *gorm.DB) repository.AuthCodeRepository {
	return &authCodeRepository{db: db}
}

// Create は新しい認可コードをデータベースに挿入します
func (r *authCodeRepository) Create(ctx context.Context, authCode *domain.AuthCode) error {
	if err := authCode.Validate(); err != nil {
		return fmt.Errorf("invalid auth code: %w", err)
	}

	a := FromDomainAuthCode(authCode)
	if err := r.db.WithContext(ctx).Create(a).Error; err != nil {
		return fmt.Errorf("failed to create auth code: %w", err)
	}

	return nil
}

// GetByCode はコードで認可コード情報を取得します
func (r *authCodeRepository) GetByCode(ctx context.Context, code string) (*domain.AuthCode, error) {
	var a AuthCode
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("auth code not found: %s", code)
		}
		return nil, fmt.Errorf("failed to get auth code: %w", err)
	}

	return a.ToDomain(), nil
}

// MarkAsUsed は認可コードを使用済みにマークします
func (r *authCodeRepository) MarkAsUsed(ctx context.Context, code string) error {
	result := r.db.WithContext(ctx).Model(&AuthCode{}).Where("code = ?", code).Update("used", true)
	if result.Error != nil {
		return fmt.Errorf("failed to mark auth code as used: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("auth code not found: %s", code)
	}
	return nil
}

// DeleteExpired は期限切れの認可コードを削除します
func (r *authCodeRepository) DeleteExpired(ctx context.Context) error {
	result := r.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&AuthCode{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete expired auth codes: %w", result.Error)
	}
	return nil
}
