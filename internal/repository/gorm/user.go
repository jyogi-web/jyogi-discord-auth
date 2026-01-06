package gorm

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
	"github.com/jyogi-web/jyogi-discord-auth/internal/repository"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository は新しいGORMユーザーリポジトリを作成します
func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepository{db: db}
}

// Create は新しいユーザーをデータベースに挿入します
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	if err := user.Validate(); err != nil {
		return fmt.Errorf("invalid user: %w", err)
	}

	u := FromDomainUser(user)
	if err := r.db.WithContext(ctx).Create(u).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID はIDでユーザーを取得します
func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var u User
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return u.ToDomain(), nil
}

// GetByDiscordID はDiscord IDでユーザーを取得します
func (r *userRepository) GetByDiscordID(ctx context.Context, discordID string) (*domain.User, error) {
	var u User
	if err := r.db.WithContext(ctx).Where("discord_id = ?", discordID).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by discord_id: %w", err)
	}

	return u.ToDomain(), nil
}

// Update は既存のユーザーを更新します
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	if err := user.Validate(); err != nil {
		return fmt.Errorf("invalid user: %w", err)
	}

	u := FromDomainUser(user)

	// 更新対象のフィールドを指定して更新
	// Selectを使用しないとゼロ値（空文字など）が無視される可能性があるが、
	// Updates(struct)の挙動ではゼロ値は更新されない。
	// 今回はすべてのフィールドを上書きして問題ないか確認が必要。
	// sqlite実装では username, avatar_url, updated_at, last_login_at のみを更新している。

	result := r.db.WithContext(ctx).Model(&User{}).Where("id = ?", u.ID).Updates(map[string]interface{}{
		"username":      u.Username,
		"display_name":  u.DisplayName,
		"avatar_url":    u.AvatarURL,
		"updated_at":    u.UpdatedAt,
		"last_login_at": u.LastLoginAt,
	})

	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found: %s", user.ID)
	}

	return nil
}

// Delete はユーザーをデータベースから削除します
func (r *userRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&User{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found: %s", id)
	}

	return nil
}

// GetAll は全てのユーザーを取得します
func (r *userRepository) GetAll(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	var users []User
	query := r.db.WithContext(ctx).Order("last_login_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}

	domainUsers := make([]*domain.User, len(users))
	for i, u := range users {
		domainUsers[i] = u.ToDomain()
	}

	return domainUsers, nil
}
