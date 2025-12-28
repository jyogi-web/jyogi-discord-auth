package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
	"github.com/jyogi-web/jyogi-discord-auth/internal/repository"
)

type profileRepository struct {
	db *sql.DB
}

// NewProfileRepository は新しいSQLiteプロフィールリポジトリを作成します
func NewProfileRepository(db *sql.DB) repository.ProfileRepository {
	return &profileRepository{db: db}
}

// Create は新しいプロフィールをデータベースに挿入します
func (r *profileRepository) Create(ctx context.Context, profile *domain.Profile) error {
	if err := profile.Validate(); err != nil {
		return fmt.Errorf("invalid profile: %w", err)
	}

	query := `
		INSERT INTO profiles (id, user_id, discord_message_id, real_name, student_id, hobbies, what_to_do, comment, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		profile.ID,
		profile.UserID,
		profile.DiscordMessageID,
		profile.RealName,
		profile.StudentID,
		profile.Hobbies,
		profile.WhatToDo,
		profile.Comment,
		profile.CreatedAt,
		profile.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}

	return nil
}

// GetByID はIDでプロフィールを取得します
func (r *profileRepository) GetByID(ctx context.Context, id string) (*domain.Profile, error) {
	query := `
		SELECT id, user_id, discord_message_id, real_name, student_id, hobbies, what_to_do, comment, created_at, updated_at
		FROM profiles
		WHERE id = ?
	`

	profile := &domain.Profile{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.DiscordMessageID,
		&profile.RealName,
		&profile.StudentID,
		&profile.Hobbies,
		&profile.WhatToDo,
		&profile.Comment,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("profile not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	return profile, nil
}

// GetByUserID はユーザーIDでプロフィールを取得します
func (r *profileRepository) GetByUserID(ctx context.Context, userID string) (*domain.Profile, error) {
	query := `
		SELECT id, user_id, discord_message_id, real_name, student_id, hobbies, what_to_do, comment, created_at, updated_at
		FROM profiles
		WHERE user_id = ?
	`

	profile := &domain.Profile{}

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.DiscordMessageID,
		&profile.RealName,
		&profile.StudentID,
		&profile.Hobbies,
		&profile.WhatToDo,
		&profile.Comment,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // Return nil, nil if profile not found (not an error)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get profile by user_id: %w", err)
	}

	return profile, nil
}

// GetByMessageID はDiscordメッセージIDでプロフィールを取得します
func (r *profileRepository) GetByMessageID(ctx context.Context, messageID string) (*domain.Profile, error) {
	query := `
		SELECT id, user_id, discord_message_id, real_name, student_id, hobbies, what_to_do, comment, created_at, updated_at
		FROM profiles
		WHERE discord_message_id = ?
	`

	profile := &domain.Profile{}

	err := r.db.QueryRowContext(ctx, query, messageID).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.DiscordMessageID,
		&profile.RealName,
		&profile.StudentID,
		&profile.Hobbies,
		&profile.WhatToDo,
		&profile.Comment,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // Return nil, nil if profile not found (not an error)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get profile by message_id: %w", err)
	}

	return profile, nil
}

// GetAll は全てのプロフィールを取得します
func (r *profileRepository) GetAll(ctx context.Context) ([]*domain.Profile, error) {
	query := `
		SELECT id, user_id, discord_message_id, real_name, student_id, hobbies, what_to_do, comment, created_at, updated_at
		FROM profiles
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all profiles: %w", err)
	}
	defer rows.Close()

	var profiles []*domain.Profile

	for rows.Next() {
		profile := &domain.Profile{}
		err := rows.Scan(
			&profile.ID,
			&profile.UserID,
			&profile.DiscordMessageID,
			&profile.RealName,
			&profile.StudentID,
			&profile.Hobbies,
			&profile.WhatToDo,
			&profile.Comment,
			&profile.CreatedAt,
			&profile.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan profile: %w", err)
		}
		profiles = append(profiles, profile)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate profiles: %w", err)
	}

	return profiles, nil
}

// Update は既存のプロフィールを更新します
func (r *profileRepository) Update(ctx context.Context, profile *domain.Profile) error {
	if err := profile.Validate(); err != nil {
		return fmt.Errorf("invalid profile: %w", err)
	}

	profile.UpdatedAt = time.Now()

	query := `
		UPDATE profiles
		SET real_name = ?, student_id = ?, hobbies = ?, what_to_do = ?, comment = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		profile.RealName,
		profile.StudentID,
		profile.Hobbies,
		profile.WhatToDo,
		profile.Comment,
		profile.UpdatedAt,
		profile.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("profile not found: %s", profile.ID)
	}

	return nil
}

// Upsert は既存のプロフィールを更新、なければ作成します
func (r *profileRepository) Upsert(ctx context.Context, profile *domain.Profile) error {
	if err := profile.Validate(); err != nil {
		return fmt.Errorf("invalid profile: %w", err)
	}

	query := `
		INSERT INTO profiles (id, user_id, discord_message_id, real_name, student_id, hobbies, what_to_do, comment, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(discord_message_id) DO UPDATE SET
			real_name = excluded.real_name,
			student_id = excluded.student_id,
			hobbies = excluded.hobbies,
			what_to_do = excluded.what_to_do,
			comment = excluded.comment,
			updated_at = excluded.updated_at
	`

	_, err := r.db.ExecContext(ctx, query,
		profile.ID,
		profile.UserID,
		profile.DiscordMessageID,
		profile.RealName,
		profile.StudentID,
		profile.Hobbies,
		profile.WhatToDo,
		profile.Comment,
		profile.CreatedAt,
		profile.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert profile: %w", err)
	}

	return nil
}

// Delete はプロフィールをデータベースから削除します
func (r *profileRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM profiles WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("profile not found: %s", id)
	}

	return nil
}
