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

// setupClientTestDB はクライアントテスト用のインメモリGORMデータベースをセットアップします
func setupClientTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	if err := db.AutoMigrate(&ClientApp{}); err != nil {
		t.Fatalf("Failed to migrate test schema: %v", err)
	}

	return db
}

// TestClientRepository_Create はクライアント作成機能をテストします
func TestClientRepository_Create(t *testing.T) {
	db := setupClientTestDB(t)
	repo := NewClientRepository(db)
	ctx := context.Background()

	client := &domain.ClientApp{
		ID:           "test-client-1",
		ClientID:     "client-id-1",
		ClientSecret: "hashed-secret-1",
		Name:         "Test App 1",
		RedirectURIs: []string{"https://example.com/callback"},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := repo.Create(ctx, client); err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// 作成されたクライアントを取得
	retrieved, err := repo.GetByID(ctx, client.ID)
	if err != nil {
		t.Fatalf("Failed to get created client: %v", err)
	}

	if retrieved.Name != client.Name {
		t.Errorf("Expected name %s, got %s", client.Name, retrieved.Name)
	}

	if len(retrieved.RedirectURIs) != 1 {
		t.Errorf("Expected 1 redirect URI, got %d", len(retrieved.RedirectURIs))
	}
}

// TestClientRepository_GetByClientID はClientIDでクライアント取得をテストします
func TestClientRepository_GetByClientID(t *testing.T) {
	db := setupClientTestDB(t)
	repo := NewClientRepository(db)
	ctx := context.Background()

	client := &domain.ClientApp{
		ID:           "test-client-2",
		ClientID:     "client-id-2",
		ClientSecret: "hashed-secret-2",
		Name:         "Test App 2",
		RedirectURIs: []string{"https://example.com/callback"},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := repo.Create(ctx, client); err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// ClientIDで取得
	retrieved, err := repo.GetByClientID(ctx, "client-id-2")
	if err != nil {
		t.Fatalf("Failed to get client by client_id: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected client, got nil")
	}

	if retrieved.ID != client.ID {
		t.Errorf("Expected client ID %s, got %s", client.ID, retrieved.ID)
	}
}

// TestClientRepository_Update はクライアント更新機能をテストします
func TestClientRepository_Update(t *testing.T) {
	db := setupClientTestDB(t)
	repo := NewClientRepository(db)
	ctx := context.Background()

	client := &domain.ClientApp{
		ID:           "test-client-3",
		ClientID:     "client-id-3",
		ClientSecret: "hashed-secret-3",
		Name:         "Test App 3",
		RedirectURIs: []string{"https://example.com/callback"},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := repo.Create(ctx, client); err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// クライアント情報を更新
	client.Name = "Updated Test App 3"
	client.RedirectURIs = []string{
		"https://example.com/callback",
		"https://example.com/callback2",
	}

	if err := repo.Update(ctx, client); err != nil {
		t.Fatalf("Failed to update client: %v", err)
	}

	// 更新されたクライアントを取得
	retrieved, err := repo.GetByID(ctx, client.ID)
	if err != nil {
		t.Fatalf("Failed to get updated client: %v", err)
	}

	if retrieved.Name != "Updated Test App 3" {
		t.Errorf("Expected name %s, got %s", "Updated Test App 3", retrieved.Name)
	}

	if len(retrieved.RedirectURIs) != 2 {
		t.Errorf("Expected 2 redirect URIs, got %d", len(retrieved.RedirectURIs))
	}
}

// TestClientRepository_GetAll は全クライアント取得をテストします
func TestClientRepository_GetAll(t *testing.T) {
	db := setupClientTestDB(t)
	repo := NewClientRepository(db)
	ctx := context.Background()

	// 複数のクライアントを作成
	clients := []*domain.ClientApp{
		{
			ID:           "client-1",
			ClientID:     "cid-1",
			ClientSecret: "secret-1",
			Name:         "App 1",
			RedirectURIs: []string{"https://app1.com/callback"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           "client-2",
			ClientID:     "cid-2",
			ClientSecret: "secret-2",
			Name:         "App 2",
			RedirectURIs: []string{"https://app2.com/callback"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           "client-3",
			ClientID:     "cid-3",
			ClientSecret: "secret-3",
			Name:         "App 3",
			RedirectURIs: []string{"https://app3.com/callback"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	for _, client := range clients {
		if err := repo.Create(ctx, client); err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
	}

	// 全クライアントを取得
	allClients, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("Failed to get all clients: %v", err)
	}

	if len(allClients) != 3 {
		t.Errorf("Expected 3 clients, got %d", len(allClients))
	}
}

// TestClientRepository_MultipleRedirectURIs は複数のリダイレクトURIをテストします
func TestClientRepository_MultipleRedirectURIs(t *testing.T) {
	db := setupClientTestDB(t)
	repo := NewClientRepository(db)
	ctx := context.Background()

	client := &domain.ClientApp{
		ID:           "test-client-multi",
		ClientID:     "client-multi",
		ClientSecret: "secret-multi",
		Name:         "Multi URI App",
		RedirectURIs: []string{
			"https://app.com/callback1",
			"https://app.com/callback2",
			"https://app.com/callback3",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := repo.Create(ctx, client); err != nil {
		t.Fatalf("Failed to create client with multiple redirect URIs: %v", err)
	}

	// 取得して確認
	retrieved, err := repo.GetByClientID(ctx, "client-multi")
	if err != nil {
		t.Fatalf("Failed to get client: %v", err)
	}

	if len(retrieved.RedirectURIs) != 3 {
		t.Errorf("Expected 3 redirect URIs, got %d", len(retrieved.RedirectURIs))
	}

	// URIの内容を確認
	expectedURIs := map[string]bool{
		"https://app.com/callback1": true,
		"https://app.com/callback2": true,
		"https://app.com/callback3": true,
	}

	for _, uri := range retrieved.RedirectURIs {
		if !expectedURIs[uri] {
			t.Errorf("Unexpected redirect URI: %s", uri)
		}
	}
}

// TestClientRepository_UniqueConstraint はClientIDの一意性制約をテストします
func TestClientRepository_UniqueConstraint(t *testing.T) {
	db := setupClientTestDB(t)
	repo := NewClientRepository(db)
	ctx := context.Background()

	client1 := &domain.ClientApp{
		ID:           "client-unique-1",
		ClientID:     "unique-client-id",
		ClientSecret: "secret-1",
		Name:         "App 1",
		RedirectURIs: []string{"https://app1.com/callback"},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := repo.Create(ctx, client1); err != nil {
		t.Fatalf("Failed to create first client: %v", err)
	}

	// 同じClientIDで別のクライアントを作成しようとする
	client2 := &domain.ClientApp{
		ID:           "client-unique-2",
		ClientID:     "unique-client-id", // 重複
		ClientSecret: "secret-2",
		Name:         "App 2",
		RedirectURIs: []string{"https://app2.com/callback"},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(ctx, client2)
	if err == nil {
		t.Error("Expected error when creating client with duplicate client_id, got nil")
	}
}
