package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
	"github.com/jyogi-web/jyogi-discord-auth/internal/repository"
)

type userRepository struct {
	db *sql.DB
}

// NewUserRepository は新しいSQLiteユーザーリポジトリを作成します
func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &userRepository{db: db}
}

// Create は新しいユーザーをデータベースに挿入します
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	if err := user.Validate(); err != nil {
		return fmt.Errorf("invalid user: %w", err)
	}

	query := `
		INSERT INTO users (id, discord_id, username, avatar_url, created_at, updated_at, last_login_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.DiscordID,
		user.Username,
		user.AvatarURL,
		user.CreatedAt,
		user.UpdatedAt,
		user.LastLoginAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID はIDでユーザーを取得します
func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, discord_id, username, avatar_url, created_at, updated_at, last_login_at
		FROM users
		WHERE id = ?
	`

	user := &domain.User{}
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.DiscordID,
		&user.Username,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// GetByDiscordID はDiscord IDでユーザーを取得します
func (r *userRepository) GetByDiscordID(ctx context.Context, discordID string) (*domain.User, error) {
	query := `
		SELECT id, discord_id, username, avatar_url, created_at, updated_at, last_login_at
		FROM users
		WHERE discord_id = ?
	`

	user := &domain.User{}
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, discordID).Scan(
		&user.ID,
		&user.DiscordID,
		&user.Username,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // Return nil, nil if user not found (not an error)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by discord_id: %w", err)
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// Update は既存のユーザーを更新します
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	if err := user.Validate(); err != nil {
		return fmt.Errorf("invalid user: %w", err)
	}

	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET username = ?, avatar_url = ?, updated_at = ?, last_login_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		user.Username,
		user.AvatarURL,
		user.UpdatedAt,
		user.LastLoginAt,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found: %s", user.ID)
	}

	return nil
}

// Delete はユーザーをデータベースから削除します
func (r *userRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found: %s", id)
	}

	return nil
}
