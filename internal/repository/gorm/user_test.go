package gorm

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
)

// setupUserTestDB はユーザーテスト用のインメモリGORMデータベースをセットアップします
func setupUserTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// AutoMigrateでテーブルを作成
	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("Failed to migrate test schema: %v", err)
	}

	return db
}

// TestUserRepository_Create はユーザー作成機能をテストします
func TestUserRepository_Create(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		ID:        "test-user-1",
		DiscordID: "discord-123",
		Username:  "testuser",
		AvatarURL: "https://example.com/avatar.png",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// 作成されたユーザーを取得
	retrieved, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}

	if retrieved.Username != user.Username {
		t.Errorf("Expected username %s, got %s", user.Username, retrieved.Username)
	}

	if retrieved.DiscordID != user.DiscordID {
		t.Errorf("Expected discord_id %s, got %s", user.DiscordID, retrieved.DiscordID)
	}
}

// TestUserRepository_GetByDiscordID はDiscordIDでユーザー取得をテストします
func TestUserRepository_GetByDiscordID(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		ID:        "test-user-2",
		DiscordID: "discord-456",
		Username:  "testuser2",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// DiscordIDで取得
	retrieved, err := repo.GetByDiscordID(ctx, "discord-456")
	if err != nil {
		t.Fatalf("Failed to get user by discord_id: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected user, got nil")
	}

	if retrieved.ID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, retrieved.ID)
	}
}

// TestUserRepository_GetByDiscordID_NotFound は存在しないDiscordIDでユーザー取得をテストします
// Medium Issue #7: SQLite実装との互換性（nil, nilを返すべき）
func TestUserRepository_GetByDiscordID_NotFound(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// 存在しないDiscordIDで取得
	retrieved, err := repo.GetByDiscordID(ctx, "non-existent-discord-id")

	// SQLite実装との互換性（nil, nilを返すべき）
	if err != nil {
		t.Errorf("Expected nil error for non-existent discord_id, got: %v", err)
	}

	if retrieved != nil {
		t.Errorf("Expected nil user for non-existent discord_id, got: %+v", retrieved)
	}
}

// TestUserRepository_Update はユーザー更新機能をテストします
func TestUserRepository_Update(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		ID:        "test-user-3",
		DiscordID: "discord-789",
		Username:  "testuser3",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// ユーザー情報を更新
	user.Username = "updateduser"
	user.AvatarURL = "https://example.com/new-avatar.png"
	loginTime := time.Now()
	user.LastLoginAt = &loginTime

	if err := repo.Update(ctx, user); err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	// 更新されたユーザーを取得
	retrieved, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	if retrieved.Username != "updateduser" {
		t.Errorf("Expected username %s, got %s", "updateduser", retrieved.Username)
	}

	if retrieved.AvatarURL != "https://example.com/new-avatar.png" {
		t.Errorf("Expected avatar_url %s, got %s", "https://example.com/new-avatar.png", retrieved.AvatarURL)
	}

	if retrieved.LastLoginAt == nil {
		t.Error("Expected last_login_at to be set")
	}
}

// TestUserRepository_Delete はユーザー削除機能をテストします
func TestUserRepository_Delete(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		ID:        "test-user-4",
		DiscordID: "discord-101",
		Username:  "testuser4",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// ユーザーを削除
	if err := repo.Delete(ctx, user.ID); err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// 削除されたことを確認
	retrieved, err := repo.GetByID(ctx, user.ID)
	if err == nil {
		t.Error("Expected error when getting deleted user, got nil")
	}

	if retrieved != nil {
		t.Errorf("Expected nil user after deletion, got: %+v", retrieved)
	}
}

// TestUserRepository_UniqueConstraint はDiscordIDの一意性制約をテストします
func TestUserRepository_UniqueConstraint(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user1 := &domain.User{
		ID:        "test-user-5",
		DiscordID: "discord-unique",
		Username:  "testuser5",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := repo.Create(ctx, user1); err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	// 同じDiscordIDで別のユーザーを作成しようとする
	user2 := &domain.User{
		ID:        "test-user-6",
		DiscordID: "discord-unique", // 重複
		Username:  "testuser6",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, user2)
	if err == nil {
		t.Error("Expected error when creating user with duplicate discord_id, got nil")
	}
}
