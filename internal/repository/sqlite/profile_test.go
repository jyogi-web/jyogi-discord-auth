package sqlite

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"

	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
)

// setupProfileTestDB はテスト用のインメモリSQLiteデータベースをセットアップします
func setupProfileTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// usersテーブルを作成（profilesの外部キー制約のため）
	_, err = db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			discord_id TEXT UNIQUE NOT NULL,
			username TEXT NOT NULL,
			avatar_url TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// profilesテーブルを作成
	_, err = db.Exec(`
		CREATE TABLE profiles (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			discord_message_id TEXT UNIQUE NOT NULL,
			real_name TEXT,
			student_id TEXT,
			hobbies TEXT,
			what_to_do TEXT,
			comment TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create profiles table: %v", err)
	}

	// インデックスを作成
	_, err = db.Exec(`CREATE INDEX idx_profiles_user_id ON profiles(user_id)`)
	if err != nil {
		t.Fatalf("Failed to create user_id index: %v", err)
	}

	_, err = db.Exec(`CREATE INDEX idx_profiles_discord_message_id ON profiles(discord_message_id)`)
	if err != nil {
		t.Fatalf("Failed to create discord_message_id index: %v", err)
	}

	return db
}

// createTestUser はテスト用のユーザーを作成します
func createTestUser(t *testing.T, db *sql.DB) *domain.User {
	user := &domain.User{
		ID:        uuid.New().String(),
		DiscordID: "test-discord-id-" + uuid.New().String(),
		Username:  "test-user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := db.Exec(`
		INSERT INTO users (id, discord_id, username, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, user.ID, user.DiscordID, user.Username, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

func TestProfileRepository_Create(t *testing.T) {
	db := setupProfileTestDB(t)
	defer db.Close()

	user := createTestUser(t, db)
	repo := NewProfileRepository(db)
	ctx := context.Background()

	profile := &domain.Profile{
		ID:               uuid.New().String(),
		UserID:           user.ID,
		DiscordMessageID: "msg-12345",
		RealName:         "田中太郎",
		StudentID:        "B1234567",
		Hobbies:          "プログラミング、読書",
		WhatToDo:         "Web開発を学びたい",
		Comment:          "よろしくお願いします",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// プロフィールを作成
	err := repo.Create(ctx, profile)
	if err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	// 作成されたプロフィールを取得
	retrieved, err := repo.GetByID(ctx, profile.ID)
	if err != nil {
		t.Fatalf("Failed to get created profile: %v", err)
	}

	// 値を確認
	if retrieved.ID != profile.ID {
		t.Errorf("ID mismatch: got %v, want %v", retrieved.ID, profile.ID)
	}
	if retrieved.UserID != profile.UserID {
		t.Errorf("UserID mismatch: got %v, want %v", retrieved.UserID, profile.UserID)
	}
	if retrieved.DiscordMessageID != profile.DiscordMessageID {
		t.Errorf("DiscordMessageID mismatch: got %v, want %v", retrieved.DiscordMessageID, profile.DiscordMessageID)
	}
	if retrieved.RealName != profile.RealName {
		t.Errorf("RealName mismatch: got %v, want %v", retrieved.RealName, profile.RealName)
	}
	if retrieved.StudentID != profile.StudentID {
		t.Errorf("StudentID mismatch: got %v, want %v", retrieved.StudentID, profile.StudentID)
	}
}

func TestProfileRepository_GetByUserID(t *testing.T) {
	db := setupProfileTestDB(t)
	defer db.Close()

	user := createTestUser(t, db)
	repo := NewProfileRepository(db)
	ctx := context.Background()

	profile := &domain.Profile{
		ID:               uuid.New().String(),
		UserID:           user.ID,
		DiscordMessageID: "msg-67890",
		RealName:         "鈴木花子",
		StudentID:        "B9876543",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// プロフィールを作成
	err := repo.Create(ctx, profile)
	if err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	// ユーザーIDで取得
	retrieved, err := repo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get profile by user_id: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected profile, got nil")
	}

	if retrieved.UserID != user.ID {
		t.Errorf("UserID mismatch: got %v, want %v", retrieved.UserID, user.ID)
	}
}

func TestProfileRepository_GetByUserID_NotFound(t *testing.T) {
	db := setupProfileTestDB(t)
	defer db.Close()

	repo := NewProfileRepository(db)
	ctx := context.Background()

	// 存在しないユーザーIDで取得
	retrieved, err := repo.GetByUserID(ctx, "non-existent-user-id")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if retrieved != nil {
		t.Error("Expected nil profile for non-existent user, got profile")
	}
}

func TestProfileRepository_GetByMessageID(t *testing.T) {
	db := setupProfileTestDB(t)
	defer db.Close()

	user := createTestUser(t, db)
	repo := NewProfileRepository(db)
	ctx := context.Background()

	messageID := "msg-unique-12345"
	profile := &domain.Profile{
		ID:               uuid.New().String(),
		UserID:           user.ID,
		DiscordMessageID: messageID,
		RealName:         "佐藤次郎",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// プロフィールを作成
	err := repo.Create(ctx, profile)
	if err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	// メッセージIDで取得
	retrieved, err := repo.GetByMessageID(ctx, messageID)
	if err != nil {
		t.Fatalf("Failed to get profile by message_id: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected profile, got nil")
	}

	if retrieved.DiscordMessageID != messageID {
		t.Errorf("DiscordMessageID mismatch: got %v, want %v", retrieved.DiscordMessageID, messageID)
	}
}

func TestProfileRepository_GetAll(t *testing.T) {
	db := setupProfileTestDB(t)
	defer db.Close()

	user := createTestUser(t, db)
	repo := NewProfileRepository(db)
	ctx := context.Background()

	// 複数のプロフィールを作成
	for i := 0; i < 3; i++ {
		profile := &domain.Profile{
			ID:               uuid.New().String(),
			UserID:           user.ID,
			DiscordMessageID: "msg-" + uuid.New().String(),
			RealName:         "テストユーザー",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		err := repo.Create(ctx, profile)
		if err != nil {
			t.Fatalf("Failed to create profile %d: %v", i, err)
		}
		time.Sleep(1 * time.Millisecond) // created_atの順序を保証
	}

	// 全プロフィールを取得
	profiles, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("Failed to get all profiles: %v", err)
	}

	if len(profiles) != 3 {
		t.Errorf("Expected 3 profiles, got %d", len(profiles))
	}
}

func TestProfileRepository_Update(t *testing.T) {
	db := setupProfileTestDB(t)
	defer db.Close()

	user := createTestUser(t, db)
	repo := NewProfileRepository(db)
	ctx := context.Background()

	profile := &domain.Profile{
		ID:               uuid.New().String(),
		UserID:           user.ID,
		DiscordMessageID: "msg-update-test",
		RealName:         "元の名前",
		StudentID:        "B1111111",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// プロフィールを作成
	err := repo.Create(ctx, profile)
	if err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	// プロフィールを更新
	profile.RealName = "更新後の名前"
	profile.StudentID = "B9999999"
	profile.Hobbies = "新しい趣味"

	err = repo.Update(ctx, profile)
	if err != nil {
		t.Fatalf("Failed to update profile: %v", err)
	}

	// 更新されたプロフィールを取得
	updated, err := repo.GetByID(ctx, profile.ID)
	if err != nil {
		t.Fatalf("Failed to get updated profile: %v", err)
	}

	if updated.RealName != "更新後の名前" {
		t.Errorf("RealName not updated: got %v, want %v", updated.RealName, "更新後の名前")
	}
	if updated.StudentID != "B9999999" {
		t.Errorf("StudentID not updated: got %v, want %v", updated.StudentID, "B9999999")
	}
	if updated.Hobbies != "新しい趣味" {
		t.Errorf("Hobbies not updated: got %v, want %v", updated.Hobbies, "新しい趣味")
	}
}

func TestProfileRepository_Upsert_Insert(t *testing.T) {
	db := setupProfileTestDB(t)
	defer db.Close()

	user := createTestUser(t, db)
	repo := NewProfileRepository(db)
	ctx := context.Background()

	messageID := "msg-upsert-new"
	profile := &domain.Profile{
		ID:               uuid.New().String(),
		UserID:           user.ID,
		DiscordMessageID: messageID,
		RealName:         "新規プロフィール",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Upsertで新規作成
	err := repo.Upsert(ctx, profile)
	if err != nil {
		t.Fatalf("Failed to upsert (insert) profile: %v", err)
	}

	// 作成されたか確認
	retrieved, err := repo.GetByMessageID(ctx, messageID)
	if err != nil {
		t.Fatalf("Failed to get upserted profile: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected profile, got nil")
	}

	if retrieved.RealName != "新規プロフィール" {
		t.Errorf("RealName mismatch: got %v, want %v", retrieved.RealName, "新規プロフィール")
	}
}

func TestProfileRepository_Upsert_Update(t *testing.T) {
	db := setupProfileTestDB(t)
	defer db.Close()

	user := createTestUser(t, db)
	repo := NewProfileRepository(db)
	ctx := context.Background()

	messageID := "msg-upsert-existing"
	profile := &domain.Profile{
		ID:               uuid.New().String(),
		UserID:           user.ID,
		DiscordMessageID: messageID,
		RealName:         "元のプロフィール",
		Hobbies:          "元の趣味",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// 最初のUpsert（新規作成）
	err := repo.Upsert(ctx, profile)
	if err != nil {
		t.Fatalf("Failed to upsert (first time): %v", err)
	}

	// 同じメッセージIDで異なる内容をUpsert（更新）
	updatedProfile := &domain.Profile{
		ID:               uuid.New().String(), // 異なるID
		UserID:           user.ID,
		DiscordMessageID: messageID, // 同じメッセージID
		RealName:         "更新されたプロフィール",
		Hobbies:          "更新された趣味",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err = repo.Upsert(ctx, updatedProfile)
	if err != nil {
		t.Fatalf("Failed to upsert (update): %v", err)
	}

	// 更新されたか確認
	retrieved, err := repo.GetByMessageID(ctx, messageID)
	if err != nil {
		t.Fatalf("Failed to get upserted profile: %v", err)
	}

	if retrieved.RealName != "更新されたプロフィール" {
		t.Errorf("RealName not updated: got %v, want %v", retrieved.RealName, "更新されたプロフィール")
	}
	if retrieved.Hobbies != "更新された趣味" {
		t.Errorf("Hobbies not updated: got %v, want %v", retrieved.Hobbies, "更新された趣味")
	}
}

func TestProfileRepository_Delete(t *testing.T) {
	db := setupProfileTestDB(t)
	defer db.Close()

	user := createTestUser(t, db)
	repo := NewProfileRepository(db)
	ctx := context.Background()

	profile := &domain.Profile{
		ID:               uuid.New().String(),
		UserID:           user.ID,
		DiscordMessageID: "msg-delete-test",
		RealName:         "削除テスト",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// プロフィールを作成
	err := repo.Create(ctx, profile)
	if err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	// プロフィールを削除
	err = repo.Delete(ctx, profile.ID)
	if err != nil {
		t.Fatalf("Failed to delete profile: %v", err)
	}

	// 削除されたか確認
	_, err = repo.GetByID(ctx, profile.ID)
	if err == nil {
		t.Error("Expected error for deleted profile, got nil")
	}
}

func TestProfileRepository_Delete_NotFound(t *testing.T) {
	db := setupProfileTestDB(t)
	defer db.Close()

	repo := NewProfileRepository(db)
	ctx := context.Background()

	// 存在しないIDで削除を試みる
	err := repo.Delete(ctx, "non-existent-profile-id")
	if err == nil {
		t.Error("Expected error for non-existent profile, got nil")
	}
}
