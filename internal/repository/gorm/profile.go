package gorm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
	"github.com/jyogi-web/jyogi-discord-auth/internal/repository"
)

type profileRepository struct {
	db *gorm.DB
}

// NewProfileRepository は新しいGORMプロフィールリポジトリを作成します
func NewProfileRepository(db *gorm.DB) repository.ProfileRepository {
	return &profileRepository{db: db}
}

// Create は新しいプロフィールをデータベースに挿入します
func (r *profileRepository) Create(ctx context.Context, profile *domain.Profile) error {
	if err := profile.Validate(); err != nil {
		return fmt.Errorf("invalid profile: %w", err)
	}

	p := FromDomainProfile(profile)
	if err := r.db.WithContext(ctx).Create(p).Error; err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}

	return nil
}

// GetByID はIDでプロフィールを取得します
func (r *profileRepository) GetByID(ctx context.Context, id string) (*domain.Profile, error) {
	var p Profile
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("profile not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	return p.ToDomain(), nil
}

// GetByUserID はユーザーIDでプロフィールを取得します
func (r *profileRepository) GetByUserID(ctx context.Context, userID string) (*domain.Profile, error) {
	var p Profile
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: user_id=%s", domain.ErrProfileNotFound, userID)
		}
		return nil, fmt.Errorf("failed to get profile by user_id: %w", err)
	}

	return p.ToDomain(), nil
}

// GetByMessageID はDiscordメッセージIDでプロフィールを取得します
func (r *profileRepository) GetByMessageID(ctx context.Context, messageID string) (*domain.Profile, error) {
	var p Profile
	if err := r.db.WithContext(ctx).Where("discord_message_id = ?", messageID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("profile not found regarding message: %s", messageID)
		}
		return nil, fmt.Errorf("failed to get profile by message_id: %w", err)
	}

	return p.ToDomain(), nil
}

// GetByUserIDs はユーザーIDのリストでプロフィールを一括取得します
func (r *profileRepository) GetByUserIDs(ctx context.Context, userIDs []string) ([]*domain.Profile, error) {
	if len(userIDs) == 0 {
		return []*domain.Profile{}, nil
	}

	var profiles []Profile
	if err := r.db.WithContext(ctx).Where("user_id IN ?", userIDs).Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("failed to get profiles by user_ids: %w", err)
	}

	domainProfiles := make([]*domain.Profile, len(profiles))
	for i, p := range profiles {
		domainProfiles[i] = p.ToDomain()
	}

	return domainProfiles, nil
}

// GetAll はすべてのプロフィールを取得します
func (r *profileRepository) GetAll(ctx context.Context) ([]*domain.Profile, error) {
	var profiles []Profile
	if err := r.db.WithContext(ctx).Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("failed to get all profiles: %w", err)
	}

	domainProfiles := make([]*domain.Profile, len(profiles))
	for i, p := range profiles {
		domainProfiles[i] = p.ToDomain()
	}

	return domainProfiles, nil
}

// Update は既存のプロフィールを更新します
func (r *profileRepository) Update(ctx context.Context, profile *domain.Profile) error {
	if err := profile.Validate(); err != nil {
		return fmt.Errorf("invalid profile: %w", err)
	}

	p := FromDomainProfile(profile)
	p.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&Profile{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
		"real_name":  p.RealName,
		"student_id": p.StudentID,
		"hobbies":    p.Hobbies,
		"what_to_do": p.WhatToDo,
		"comment":    p.Comment,
		"updated_at": p.UpdatedAt,
	})

	if result.Error != nil {
		return fmt.Errorf("failed to update profile: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("profile not found: %s", profile.ID)
	}

	return nil
}

// Upsert はプロフィールを挿入または更新します
func (r *profileRepository) Upsert(ctx context.Context, profile *domain.Profile) error {
	if err := profile.Validate(); err != nil {
		return fmt.Errorf("invalid profile: %w", err)
	}

	p := FromDomainProfile(profile)

	// Conflict対象のカラム(Constraint)を指定する必要があるが、
	// MySQL/GORMではOnConflictを使用してDuplicate Key Updateを行う
	// IDまたはUnique Index (discord_message_id) での競合を想定

	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "discord_message_id"}}, // Unique constraint
		UpdateAll: true,
	}).Create(p).Error

	if err != nil {
		return fmt.Errorf("failed to upsert profile: %w", err)
	}

	return nil
}

// Delete はプロフィールをデータベースから削除します
func (r *profileRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&Profile{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete profile: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("profile not found: %s", id)
	}

	return nil
}
