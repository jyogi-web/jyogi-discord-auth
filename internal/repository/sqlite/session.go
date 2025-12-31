package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
	"github.com/jyogi-web/jyogi-discord-auth/internal/repository"
)

type sessionRepository struct {
	db *sql.DB
}

// NewSessionRepository は新しいSQLiteセッションリポジトリを作成します
func NewSessionRepository(db *sql.DB) repository.SessionRepository {
	return &sessionRepository{db: db}
}

// Create は新しいセッションをデータベースに挿入します
func (r *sessionRepository) Create(ctx context.Context, session *domain.Session) error {
	if err := session.Validate(); err != nil {
		return fmt.Errorf("invalid session: %w", err)
	}

	query := `
		INSERT INTO sessions (id, user_id, token, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.UserID,
		session.Token,
		session.ExpiresAt.Format(time.RFC3339),
		session.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetByID はIDでセッションを取得します
func (r *sessionRepository) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM sessions
		WHERE id = ?
	`

	var expiresAtStr, createdAtStr string
	session := &domain.Session{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID,
		&session.UserID,
		&session.Token,
		&expiresAtStr,
		&createdAtStr,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// 日付文字列をパース
	session.ExpiresAt, err = time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expires_at: %w", err)
	}
	session.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	return session, nil
}

// GetByToken はトークンでセッションを取得します
func (r *sessionRepository) GetByToken(ctx context.Context, token string) (*domain.Session, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM sessions
		WHERE token = ? AND expires_at > ?
	`

	var expiresAtStr, createdAtStr string
	session := &domain.Session{}
	err := r.db.QueryRowContext(ctx, query, token, time.Now().Format(time.RFC3339)).Scan(
		&session.ID,
		&session.UserID,
		&session.Token,
		&expiresAtStr,
		&createdAtStr,
	)
	if err == sql.ErrNoRows {
		return nil, nil // Return nil, nil if session not found or expired
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session by token: %w", err)
	}

	// 日付文字列をパース
	session.ExpiresAt, err = time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expires_at: %w", err)
	}
	session.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	return session, nil
}

// GetByUserID はユーザーの全セッションを取得します
func (r *sessionRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Session, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM sessions
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions by user_id: %w", err)
	}
	defer rows.Close()

	sessions := []*domain.Session{}
	for rows.Next() {
		var expiresAtStr, createdAtStr string
		session := &domain.Session{}
		if err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.Token,
			&expiresAtStr,
			&createdAtStr,
		); err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		// 日付文字列をパース
		session.ExpiresAt, err = time.Parse(time.RFC3339, expiresAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse expires_at: %w", err)
		}
		session.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}

		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

// Delete はIDでセッションを削除します
func (r *sessionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM sessions WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", id)
	}

	return nil
}

// DeleteByToken はトークンでセッションを削除します
func (r *sessionRepository) DeleteByToken(ctx context.Context, token string) error {
	query := `DELETE FROM sessions WHERE token = ?`

	_, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to delete session by token: %w", err)
	}

	return nil
}

// DeleteExpired は期限切れのセッションをすべて削除します
func (r *sessionRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at < ?`

	result, err := r.db.ExecContext(ctx, query, time.Now().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	// 削除されたセッション数をログ出力（オプション）
	if rowsAffected > 0 {
		fmt.Printf("Deleted %d expired sessions\n", rowsAffected)
	}

	return nil
}
