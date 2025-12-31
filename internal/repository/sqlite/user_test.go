package sqlite

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
)

// setupTestDB はテスト用のインメモリデータベースをセットアップします
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// テーブルを作成
	schema := `
	CREATE TABLE users (
		id TEXT PRIMARY KEY,
		discord_id TEXT UNIQUE NOT NULL,
		username TEXT NOT NULL,
		avatar_url TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_login_at TIMESTAMP
	);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	return db
}

// TestUserRepository_Create はユーザー作成機能をテストします
func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		ID:        "test-user-1",
		DiscordID: "123456789",
		Username:  "testuser",
		AvatarURL: "https://example.com/avatar.png",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// ユーザーを作成
	err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// 作成したユーザーを取得して確認
	retrieved, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if retrieved.ID != user.ID {
		t.Errorf("ID = %v, want %v", retrieved.ID, user.ID)
	}

	if retrieved.DiscordID != user.DiscordID {
		t.Errorf("DiscordID = %v, want %v", retrieved.DiscordID, user.DiscordID)
	}

	if retrieved.Username != user.Username {
		t.Errorf("Username = %v, want %v", retrieved.Username, user.Username)
	}
}

// TestUserRepository_Create_ValidationError はバリデーションエラーをテストします
func TestUserRepository_Create_ValidationError(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	tests := []struct {
		name string
		user *domain.User
	}{
		{
			name: "DiscordIDが空",
			user: &domain.User{
				ID:        "test-user-1",
				DiscordID: "", // 空
				Username:  "testuser",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		{
			name: "Usernameが空",
			user: &domain.User{
				ID:        "test-user-1",
				DiscordID: "123456789",
				Username:  "", // 空
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, tt.user)
			if err == nil {
				t.Error("Expected validation error, got nil")
			}
		})
	}
}

// TestUserRepository_GetByDiscordID はDiscord IDでユーザーを取得する機能をテストします
func TestUserRepository_GetByDiscordID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		ID:        "test-user-1",
		DiscordID: "123456789",
		Username:  "testuser",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// ユーザーを作成
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Discord IDで取得
	retrieved, err := repo.GetByDiscordID(ctx, user.DiscordID)
	if err != nil {
		t.Fatalf("Failed to get user by discord_id: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected user, got nil")
	}

	if retrieved.DiscordID != user.DiscordID {
		t.Errorf("DiscordID = %v, want %v", retrieved.DiscordID, user.DiscordID)
	}

	// 存在しないDiscord ID
	notFound, err := repo.GetByDiscordID(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if notFound != nil {
		t.Error("Expected nil for nonexistent user, got user")
	}
}

// TestUserRepository_Update はユーザー更新機能をテストします
func TestUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		ID:        "test-user-1",
		DiscordID: "123456789",
		Username:  "testuser",
		AvatarURL: "https://example.com/avatar.png",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// ユーザーを作成
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// ユーザー情報を更新
	user.Username = "updateduser"
	user.AvatarURL = "https://example.com/new-avatar.png"
	now := time.Now()
	user.LastLoginAt = &now

	if err := repo.Update(ctx, user); err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	// 更新されたユーザーを取得
	retrieved, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if retrieved.Username != "updateduser" {
		t.Errorf("Username = %v, want %v", retrieved.Username, "updateduser")
	}

	if retrieved.AvatarURL != "https://example.com/new-avatar.png" {
		t.Errorf("AvatarURL = %v, want %v", retrieved.AvatarURL, "https://example.com/new-avatar.png")
	}

	if retrieved.LastLoginAt == nil {
		t.Error("LastLoginAt should not be nil")
	}
}

// TestUserRepository_Delete はユーザー削除機能をテストします
func TestUserRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		ID:        "test-user-1",
		DiscordID: "123456789",
		Username:  "testuser",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// ユーザーを作成
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// ユーザーを削除
	if err := repo.Delete(ctx, user.ID); err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// 削除されたユーザーを取得しようとする
	_, err := repo.GetByID(ctx, user.ID)
	if err == nil {
		t.Error("Expected error when getting deleted user, got nil")
	}

	// 存在しないユーザーを削除しようとする
	err = repo.Delete(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error when deleting nonexistent user, got nil")
	}
}
