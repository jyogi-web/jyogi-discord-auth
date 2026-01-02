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
// 期限切れのセッションは見つからなかったものとして扱います（SQLite実装との互換性のため）
func (r *sessionRepository) GetByToken(ctx context.Context, token string) (*domain.Session, error) {
	var s Session
	if err := r.db.WithContext(ctx).Where("token = ? AND expires_at > ?", token, time.Now()).First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// SQLite実装との互換性のため、nil, nilを返す
			return nil, nil
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
// SQLite実装との互換性のため、削除対象が存在しない場合もエラーを返しません
func (r *sessionRepository) DeleteByToken(ctx context.Context, token string) error {
	result := r.db.WithContext(ctx).Delete(&Session{}, "token = ?", token)
	if result.Error != nil {
		return fmt.Errorf("failed to delete session by token: %w", result.Error)
	}
	// SQLite実装との互換性のため、RowsAffected == 0 でもエラーを返さない
	return nil
}

// DeleteExpired は期限切れのセッションをすべて削除します
func (r *sessionRepository) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Where("expires_at < ?", now).Delete(&Session{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", result.Error)
	}

	// 削除件数をログに記録（運用性向上）
	if result.RowsAffected > 0 {
		fmt.Printf("Deleted %d expired session(s) (before %v)\n", result.RowsAffected, now)
	}

	return nil
}
