package gorm_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
	gormRepo "github.com/jyogi-web/jyogi-discord-auth/internal/repository/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&gormRepo.Profile{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

func TestProfileRepository_GetByUserID(t *testing.T) {
	db := setupTestDB(t)
	repo := gormRepo.NewProfileRepository(db)
	ctx := context.Background()

	t.Run("RecordNotFound", func(t *testing.T) {
		userID := uuid.New().String()
		_, err := repo.GetByUserID(ctx, userID)

		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		if !errors.Is(err, domain.ErrProfileNotFound) {
			t.Errorf("Expected error to be ErrProfileNotFound, got %v", err)
		}
	})

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New().String()
		profile := &domain.Profile{
			ID:               uuid.New().String(),
			UserID:           userID,
			DiscordMessageID: "123456789",
			RealName:         "Test User",
		}

		if err := repo.Create(ctx, profile); err != nil {
			t.Fatalf("Failed to create profile: %v", err)
		}

		result, err := repo.GetByUserID(ctx, userID)
		if err != nil {
			t.Fatalf("Failed to get profile: %v", err)
		}

		if result.ID != profile.ID {
			t.Errorf("Expected ID %s, got %s", profile.ID, result.ID)
		}
	})
}
