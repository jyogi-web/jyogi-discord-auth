package sqlite

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
)

// setupClientTestDB はテスト用のインメモリデータベースをセットアップします
func setupClientTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// テーブルを作成
	schema := `
	CREATE TABLE client_apps (
		id TEXT PRIMARY KEY,
		client_id TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		client_secret TEXT NOT NULL,
		redirect_uris TEXT NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	return db
}

func TestClientRepository_Create(t *testing.T) {
	db := setupClientTestDB(t)
	defer db.Close()

	repo := NewClientRepository(db)
	ctx := context.Background()

	now := time.Now()
	client := &domain.ClientApp{
		ID:           "test-client-1",
		ClientID:     "client_test_1",
		Name:         "Test Client App",
		ClientSecret: "hashed_secret",
		RedirectURIs: []string{"http://localhost:3000/callback"},
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err := repo.Create(ctx, client)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// 作成されたクライアントを取得して確認
	retrieved, err := repo.GetByID(ctx, client.ID)
	if err != nil {
		t.Fatalf("Failed to get client: %v", err)
	}

	if retrieved.ID != client.ID {
		t.Errorf("Expected ID %s, got %s", client.ID, retrieved.ID)
	}
	if retrieved.Name != client.Name {
		t.Errorf("Expected Name %s, got %s", client.Name, retrieved.Name)
	}
	if retrieved.ClientSecret != client.ClientSecret {
		t.Errorf("Expected ClientSecret %s, got %s", client.ClientSecret, retrieved.ClientSecret)
	}
}

func TestClientRepository_GetByID(t *testing.T) {
	db := setupClientTestDB(t)
	defer db.Close()

	repo := NewClientRepository(db)
	ctx := context.Background()

	// テストデータを作成
	now := time.Now()
	client := &domain.ClientApp{
		ID:           "test-client-2",
		ClientID:     "client_test_2",
		Name:         "Another Test Client",
		ClientSecret: "another_hashed_secret",
		RedirectURIs: []string{"http://localhost:4000/callback"},
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := repo.Create(ctx, client); err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// IDで取得
	retrieved, err := repo.GetByID(ctx, client.ID)
	if err != nil {
		t.Fatalf("Failed to get client by ID: %v", err)
	}

	if retrieved.ID != client.ID {
		t.Errorf("Expected ID %s, got %s", client.ID, retrieved.ID)
	}
}

func TestClientRepository_GetByID_NotFound(t *testing.T) {
	db := setupClientTestDB(t)
	defer db.Close()

	repo := NewClientRepository(db)
	ctx := context.Background()

	// 存在しないIDで取得を試みる
	_, err := repo.GetByID(ctx, "non-existent-client")
	if err == nil {
		t.Error("Expected error for non-existent client, got nil")
	}
}

func TestClientRepository_ValidateRedirectURI(t *testing.T) {
	db := setupClientTestDB(t)
	defer db.Close()

	repo := NewClientRepository(db)
	ctx := context.Background()

	// テストデータを作成
	now := time.Now()
	client := &domain.ClientApp{
		ID:           "test-client-3",
		ClientID:     "client_test_3",
		Name:         "URI Validation Test",
		ClientSecret: "secret",
		RedirectURIs: []string{
			"http://localhost:3000/callback",
			"http://localhost:3000/auth/callback",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := repo.Create(ctx, client); err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// 有効なリダイレクトURIをテスト
	testCases := []struct {
		name        string
		redirectURI string
		expectValid bool
	}{
		{"valid URI 1", "http://localhost:3000/callback", true},
		{"valid URI 2", "http://localhost:3000/auth/callback", true},
		{"invalid URI", "http://localhost:3000/invalid", false},
		{"completely different URI", "http://example.com/callback", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			valid, err := repo.ValidateRedirectURI(ctx, client.ClientID, tc.redirectURI)
			if err != nil {
				t.Fatalf("Failed to validate redirect URI: %v", err)
			}

			if valid != tc.expectValid {
				t.Errorf("Expected validation result %v for URI %s, got %v",
					tc.expectValid, tc.redirectURI, valid)
			}
		})
	}
}

func TestClientRepository_GetAll(t *testing.T) {
	db := setupClientTestDB(t)
	defer db.Close()

	repo := NewClientRepository(db)
	ctx := context.Background()

	// 複数のクライアントを作成
	now := time.Now()
	clients := []*domain.ClientApp{
		{
			ID:           "client-1",
			ClientID:     "client_1",
			Name:         "Client 1",
			ClientSecret: "secret1",
			RedirectURIs: []string{"http://localhost:3000/callback"},
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           "client-2",
			ClientID:     "client_2",
			Name:         "Client 2",
			ClientSecret: "secret2",
			RedirectURIs: []string{"http://localhost:4000/callback"},
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}

	for _, client := range clients {
		if err := repo.Create(ctx, client); err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
	}

	// すべてのクライアントを取得
	retrieved, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("Failed to get all clients: %v", err)
	}

	if len(retrieved) != len(clients) {
		t.Errorf("Expected %d clients, got %d", len(clients), len(retrieved))
	}
}
