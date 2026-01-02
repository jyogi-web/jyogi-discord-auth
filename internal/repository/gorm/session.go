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

type sessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository は新しいGORMセッションリポジトリを作成します
func NewSessionRepository(db *gorm.DB) repository.SessionRepository {
	return &sessionRepository{db: db}
}

// Create は新しいセッションをデータベースに挿入します
func (r *sessionRepository) Create(ctx context.Context, session *domain.Session) error {
	if err := session.Validate(); err != nil {
		return fmt.Errorf("invalid session: %w", err)
	}

	s := FromDomainSession(session)
	if err := r.db.WithContext(ctx).Create(s).Error; err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetByID はIDでセッションを取得します
func (r *sessionRepository) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	var s Session
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("session not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return s.ToDomain(), nil
}

// GetByToken はトークンでセッションを取得します
func (r *sessionRepository) GetByToken(ctx context.Context, token string) (*domain.Session, error) {
	var s Session
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("session not found by token")
		}
		return nil, fmt.Errorf("failed to get session by token: %w", err)
	}

	return s.ToDomain(), nil
}

// GetByUserID はユーザーIDに関連するすべてのセッションを取得します
func (r *sessionRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Session, error) {
	var sessions []Session
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("failed to get sessions by user_id: %w", err)
	}

	domainSessions := make([]*domain.Session, len(sessions))
	for i, s := range sessions {
		domainSessions[i] = s.ToDomain()
	}

	return domainSessions, nil
}

// Delete はセッションをデータベースから削除します
func (r *sessionRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&Session{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete session: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("session not found: %s", id)
	}
	return nil
}

// DeleteByToken はトークンでセッションを削除します
func (r *sessionRepository) DeleteByToken(ctx context.Context, token string) error {
	result := r.db.WithContext(ctx).Delete(&Session{}, "token = ?", token)
	if result.Error != nil {
		return fmt.Errorf("failed to delete session by token: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("session not found by token")
	}
	return nil
}

// DeleteExpired は期限切れのセッションをすべて削除します
func (r *sessionRepository) DeleteExpired(ctx context.Context) error {
	// 期限切れかつ、expires_atが現在時刻より前のレコードを削除
	result := r.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&Session{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", result.Error)
	}
	return nil
}
