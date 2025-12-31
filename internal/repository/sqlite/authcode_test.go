package sqlite

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
)

// setupAuthCodeTestDB はテスト用のインメモリデータベースをセットアップします
func setupAuthCodeTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// テーブルを作成
	schema := `
	CREATE TABLE auth_codes (
		id TEXT PRIMARY KEY,
		code TEXT NOT NULL UNIQUE,
		client_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		redirect_uri TEXT NOT NULL,
		expires_at TEXT NOT NULL,
		created_at TEXT NOT NULL,
		used INTEGER NOT NULL DEFAULT 0
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	return db
}

func TestAuthCodeRepository_Create(t *testing.T) {
	db := setupAuthCodeTestDB(t)
	defer db.Close()

	repo := NewAuthCodeRepository(db)
	ctx := context.Background()

	now := time.Now()
	authCode := &domain.AuthCode{
		ID:          "authcode-1",
		Code:        "test_authorization_code_123",
		ClientID:    "client-1",
		UserID:      "user-1",
		RedirectURI: "http://localhost:3000/callback",
		ExpiresAt:   now.Add(10 * time.Minute),
		CreatedAt:   now,
		Used:        false,
	}

	err := repo.Create(ctx, authCode)
	if err != nil {
		t.Fatalf("Failed to create auth code: %v", err)
	}

	// 作成された認可コードを取得して確認
	retrieved, err := repo.GetByCode(ctx, authCode.Code)
	if err != nil {
		t.Fatalf("Failed to get auth code: %v", err)
	}

	if retrieved.Code != authCode.Code {
		t.Errorf("Expected Code %s, got %s", authCode.Code, retrieved.Code)
	}
	if retrieved.ClientID != authCode.ClientID {
		t.Errorf("Expected ClientID %s, got %s", authCode.ClientID, retrieved.ClientID)
	}
	if retrieved.UserID != authCode.UserID {
		t.Errorf("Expected UserID %s, got %s", authCode.UserID, retrieved.UserID)
	}
	if retrieved.Used {
		t.Error("Expected Used to be false, got true")
	}
}

func TestAuthCodeRepository_GetByCode(t *testing.T) {
	db := setupAuthCodeTestDB(t)
	defer db.Close()

	repo := NewAuthCodeRepository(db)
	ctx := context.Background()

	// テストデータを作成
	now := time.Now()
	authCode := &domain.AuthCode{
		ID:          "authcode-2",
		Code:        "another_test_code_456",
		ClientID:    "client-2",
		UserID:      "user-2",
		RedirectURI: "http://localhost:4000/callback",
		ExpiresAt:   now.Add(10 * time.Minute),
		CreatedAt:   now,
		Used:        false,
	}

	if err := repo.Create(ctx, authCode); err != nil {
		t.Fatalf("Failed to create auth code: %v", err)
	}

	// コードで取得
	retrieved, err := repo.GetByCode(ctx, authCode.Code)
	if err != nil {
		t.Fatalf("Failed to get auth code by code: %v", err)
	}

	if retrieved.Code != authCode.Code {
		t.Errorf("Expected Code %s, got %s", authCode.Code, retrieved.Code)
	}
}

func TestAuthCodeRepository_GetByCode_NotFound(t *testing.T) {
	db := setupAuthCodeTestDB(t)
	defer db.Close()

	repo := NewAuthCodeRepository(db)
	ctx := context.Background()

	// 存在しないコードで取得を試みる
	_, err := repo.GetByCode(ctx, "non_existent_code")
	if err == nil {
		t.Error("Expected error for non-existent auth code, got nil")
	}
}

func TestAuthCodeRepository_MarkAsUsed(t *testing.T) {
	db := setupAuthCodeTestDB(t)
	defer db.Close()

	repo := NewAuthCodeRepository(db)
	ctx := context.Background()

	// テストデータを作成
	now := time.Now()
	authCode := &domain.AuthCode{
		ID:          "authcode-3",
		Code:        "mark_as_used_test_code",
		ClientID:    "client-3",
		UserID:      "user-3",
		RedirectURI: "http://localhost:5000/callback",
		ExpiresAt:   now.Add(10 * time.Minute),
		CreatedAt:   now,
		Used:        false,
	}

	if err := repo.Create(ctx, authCode); err != nil {
		t.Fatalf("Failed to create auth code: %v", err)
	}

	// 使用済みにマーク
	err := repo.MarkAsUsed(ctx, authCode.Code)
	if err != nil {
		t.Fatalf("Failed to mark auth code as used: %v", err)
	}

	// 使用済みフラグを確認
	retrieved, err := repo.GetByCode(ctx, authCode.Code)
	if err != nil {
		t.Fatalf("Failed to get auth code: %v", err)
	}

	if !retrieved.Used {
		t.Error("Expected Used to be true after MarkAsUsed, got false")
	}
}

func TestAuthCodeRepository_DeleteExpired(t *testing.T) {
	db := setupAuthCodeTestDB(t)
	defer db.Close()

	repo := NewAuthCodeRepository(db)
	ctx := context.Background()

	now := time.Now()

	// 期限切れの認可コードを作成
	expiredCode := &domain.AuthCode{
		ID:          "authcode-expired",
		Code:        "expired_code",
		ClientID:    "client-4",
		UserID:      "user-4",
		RedirectURI: "http://localhost:6000/callback",
		ExpiresAt:   now.Add(-1 * time.Hour), // 1時間前に期限切れ
		CreatedAt:   now.Add(-2 * time.Hour),
		Used:        false,
	}

	// 有効な認可コードを作成
	validCode := &domain.AuthCode{
		ID:          "authcode-valid",
		Code:        "valid_code",
		ClientID:    "client-5",
		UserID:      "user-5",
		RedirectURI: "http://localhost:7000/callback",
		ExpiresAt:   now.Add(10 * time.Minute), // まだ有効
		CreatedAt:   now,
		Used:        false,
	}

	if err := repo.Create(ctx, expiredCode); err != nil {
		t.Fatalf("Failed to create expired auth code: %v", err)
	}
	if err := repo.Create(ctx, validCode); err != nil {
		t.Fatalf("Failed to create valid auth code: %v", err)
	}

	// 期限切れの認可コードを削除
	err := repo.DeleteExpired(ctx)
	if err != nil {
		t.Fatalf("Failed to delete expired auth codes: %v", err)
	}

	// 期限切れのコードが削除されたことを確認
	_, err = repo.GetByCode(ctx, expiredCode.Code)
	if err == nil {
		t.Error("Expected error when getting expired code, got nil")
	}

	// 有効なコードは残っていることを確認
	retrieved, err := repo.GetByCode(ctx, validCode.Code)
	if err != nil {
		t.Fatalf("Expected valid code to still exist: %v", err)
	}
	if retrieved.Code != validCode.Code {
		t.Errorf("Expected Code %s, got %s", validCode.Code, retrieved.Code)
	}
}
