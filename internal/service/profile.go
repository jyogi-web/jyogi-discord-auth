package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
	"github.com/jyogi-web/jyogi-discord-auth/internal/repository"
	"github.com/jyogi-web/jyogi-discord-auth/pkg/discord"
)

// SyncStats はプロフィール同期の統計情報を保持します
type SyncStats struct {
	SuccessCount  int
	SkipCount     int
	ErrorCount    int
	TotalMessages int
}

// ProfileService はプロフィール同期サービスを提供します
type ProfileService struct {
	profileRepo   repository.ProfileRepository
	userRepo      repository.UserRepository
	botToken      string
	channelID     string
	lastSyncStats SyncStats
	mu            sync.RWMutex
}

// NewProfileService は新しいProfileServiceを作成します
func NewProfileService(
	profileRepo repository.ProfileRepository,
	userRepo repository.UserRepository,
	botToken string,
	channelID string,
) *ProfileService {
	return &ProfileService{
		profileRepo: profileRepo,
		userRepo:    userRepo,
		botToken:    botToken,
		channelID:   channelID,
	}
}

// SyncProfiles はDiscord自己紹介チャンネルからプロフィールを同期します
func (s *ProfileService) SyncProfiles(ctx context.Context) error {
	log.Println("Starting profile synchronization...")

	// チャンネルのすべてのメッセージを取得（ページネーション対応）
	messages, err := discord.GetAllChannelMessages(ctx, s.botToken, s.channelID, 0)
	if err != nil {
		return fmt.Errorf("failed to get channel messages: %w", err)
	}

	log.Printf("Retrieved %d messages from channel", len(messages))

	successCount := 0
	skipCount := 0
	errorCount := 0

	for _, msg := range messages {
		// メッセージからプロフィールをパース
		profileData := discord.ParseProfile(msg.Content)

		// 有効なプロフィールでない場合はスキップ
		if !profileData.IsValidProfile() {
			log.Printf("Skipping message %s: not a valid profile", msg.ID)
			skipCount++
			continue
		}

		// Discord IDでユーザーを検索または作成
		user, err := s.userRepo.GetByDiscordID(ctx, msg.Author.ID)
		if err != nil {
			log.Printf("Error getting user by discord_id %s: %v", msg.Author.ID, err)
			errorCount++
			continue
		}

		// ユーザーが存在しない場合は作成
		if user == nil {
			user = &domain.User{
				ID:        uuid.New().String(),
				DiscordID: msg.Author.ID,
				Username:  msg.Author.Username,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			if err := s.userRepo.Create(ctx, user); err != nil {
				log.Printf("Error creating user for discord_id %s: %v", msg.Author.ID, err)
				errorCount++
				continue
			}

			log.Printf("Created new user: %s (discord_id: %s)", user.Username, user.DiscordID)
		}

		// プロフィールを作成または更新
		profile := &domain.Profile{
			ID:               uuid.New().String(),
			UserID:           user.ID,
			DiscordMessageID: msg.ID,
			RealName:         profileData.RealName,
			StudentID:        profileData.StudentID,
			Hobbies:          profileData.Hobbies,
			WhatToDo:         profileData.WhatToDo,
			Comment:          profileData.Comment,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		if err := s.profileRepo.Upsert(ctx, profile); err != nil {
			log.Printf("Error upserting profile for message %s: %v", msg.ID, err)
			errorCount++
			continue
		}

		log.Printf("Synced profile for user %s (message: %s)", user.Username, msg.ID)
		successCount++
	}

	log.Printf("Profile synchronization completed: %d success, %d skipped, %d errors", successCount, skipCount, errorCount)

	// 統計情報を保存
	s.mu.Lock()
	s.lastSyncStats = SyncStats{
		SuccessCount:  successCount,
		SkipCount:     skipCount,
		ErrorCount:    errorCount,
		TotalMessages: len(messages),
	}
	s.mu.Unlock()

	return nil
}

// GetProfileByUserID はユーザーIDでプロフィールを取得します
func (s *ProfileService) GetProfileByUserID(ctx context.Context, userID string) (*domain.Profile, error) {
	return s.profileRepo.GetByUserID(ctx, userID)
}

// GetAllProfiles は全てのプロフィールを取得します
func (s *ProfileService) GetAllProfiles(ctx context.Context) ([]*domain.Profile, error) {
	return s.profileRepo.GetAll(ctx)
}

// GetLastSyncStats は最後の同期の統計情報を取得します
func (s *ProfileService) GetLastSyncStats() SyncStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastSyncStats
}
