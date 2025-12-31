package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
)

// モックProfileRepository
type mockProfileRepository struct {
	profiles       map[string]*domain.Profile
	profilesByUser map[string]*domain.Profile
	createError    error
	upsertError    error
}

func newMockProfileRepository() *mockProfileRepository {
	return &mockProfileRepository{
		profiles:       make(map[string]*domain.Profile),
		profilesByUser: make(map[string]*domain.Profile),
	}
}

func (m *mockProfileRepository) Create(ctx context.Context, profile *domain.Profile) error {
	if m.createError != nil {
		return m.createError
	}
	m.profiles[profile.ID] = profile
	m.profilesByUser[profile.UserID] = profile
	return nil
}

func (m *mockProfileRepository) GetByID(ctx context.Context, id string) (*domain.Profile, error) {
	profile, ok := m.profiles[id]
	if !ok {
		return nil, errors.New("profile not found")
	}
	return profile, nil
}

func (m *mockProfileRepository) GetByUserID(ctx context.Context, userID string) (*domain.Profile, error) {
	profile, ok := m.profilesByUser[userID]
	if !ok {
		return nil, nil
	}
	return profile, nil
}

func (m *mockProfileRepository) GetByMessageID(ctx context.Context, messageID string) (*domain.Profile, error) {
	for _, profile := range m.profiles {
		if profile.DiscordMessageID == messageID {
			return profile, nil
		}
	}
	return nil, nil
}

func (m *mockProfileRepository) GetAll(ctx context.Context) ([]*domain.Profile, error) {
	var profiles []*domain.Profile
	for _, p := range m.profiles {
		profiles = append(profiles, p)
	}
	return profiles, nil
}

func (m *mockProfileRepository) Update(ctx context.Context, profile *domain.Profile) error {
	m.profiles[profile.ID] = profile
	m.profilesByUser[profile.UserID] = profile
	return nil
}

func (m *mockProfileRepository) Upsert(ctx context.Context, profile *domain.Profile) error {
	if m.upsertError != nil {
		return m.upsertError
	}
	m.profiles[profile.ID] = profile
	m.profilesByUser[profile.UserID] = profile
	return nil
}

func (m *mockProfileRepository) Delete(ctx context.Context, id string) error {
	delete(m.profiles, id)
	return nil
}

// モックUserRepository
type mockUserRepository struct {
	users             map[string]*domain.User
	usersByDiscordID  map[string]*domain.User
	createError       error
	getByDiscordIDErr error
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users:            make(map[string]*domain.User),
		usersByDiscordID: make(map[string]*domain.User),
	}
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if m.createError != nil {
		return m.createError
	}
	m.users[user.ID] = user
	m.usersByDiscordID[user.DiscordID] = user
	return nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *mockUserRepository) GetByDiscordID(ctx context.Context, discordID string) (*domain.User, error) {
	if m.getByDiscordIDErr != nil {
		return nil, m.getByDiscordIDErr
	}
	user, ok := m.usersByDiscordID[discordID]
	if !ok {
		return nil, nil
	}
	return user, nil
}

func (m *mockUserRepository) Update(ctx context.Context, user *domain.User) error {
	m.users[user.ID] = user
	m.usersByDiscordID[user.DiscordID] = user
	return nil
}

func (m *mockUserRepository) Delete(ctx context.Context, id string) error {
	user, ok := m.users[id]
	if ok {
		delete(m.usersByDiscordID, user.DiscordID)
		delete(m.users, id)
	}
	return nil
}

func TestProfileService_GetProfileByUserID(t *testing.T) {
	profileRepo := newMockProfileRepository()
	userRepo := newMockUserRepository()

	service := NewProfileService(profileRepo, userRepo, "test-token", "test-channel")

	userID := uuid.New().String()
	expectedProfile := &domain.Profile{
		ID:               uuid.New().String(),
		UserID:           userID,
		DiscordMessageID: "msg-123",
		RealName:         "テストユーザー",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	profileRepo.profilesByUser[userID] = expectedProfile

	ctx := context.Background()
	profile, err := service.GetProfileByUserID(ctx, userID)

	if err != nil {
		t.Fatalf("GetProfileByUserID failed: %v", err)
	}

	if profile == nil {
		t.Fatal("Expected profile, got nil")
	}

	if profile.UserID != userID {
		t.Errorf("UserID mismatch: got %v, want %v", profile.UserID, userID)
	}

	if profile.RealName != expectedProfile.RealName {
		t.Errorf("RealName mismatch: got %v, want %v", profile.RealName, expectedProfile.RealName)
	}
}

func TestProfileService_GetProfileByUserID_NotFound(t *testing.T) {
	profileRepo := newMockProfileRepository()
	userRepo := newMockUserRepository()

	service := NewProfileService(profileRepo, userRepo, "test-token", "test-channel")

	ctx := context.Background()
	profile, err := service.GetProfileByUserID(ctx, "non-existent-user-id")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if profile != nil {
		t.Error("Expected nil profile for non-existent user, got profile")
	}
}

func TestProfileService_GetAllProfiles(t *testing.T) {
	profileRepo := newMockProfileRepository()
	userRepo := newMockUserRepository()

	service := NewProfileService(profileRepo, userRepo, "test-token", "test-channel")

	// 複数のプロフィールを追加
	for i := 0; i < 3; i++ {
		profile := &domain.Profile{
			ID:               uuid.New().String(),
			UserID:           uuid.New().String(),
			DiscordMessageID: "msg-" + uuid.New().String(),
			RealName:         "テストユーザー",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		profileRepo.profiles[profile.ID] = profile
	}

	ctx := context.Background()
	profiles, err := service.GetAllProfiles(ctx)

	if err != nil {
		t.Fatalf("GetAllProfiles failed: %v", err)
	}

	if len(profiles) != 3 {
		t.Errorf("Expected 3 profiles, got %d", len(profiles))
	}
}

func TestProfileService_GetLastSyncStats(t *testing.T) {
	profileRepo := newMockProfileRepository()
	userRepo := newMockUserRepository()

	service := NewProfileService(profileRepo, userRepo, "test-token", "test-channel")

	// 初期状態のstatsを確認
	stats := service.GetLastSyncStats()

	if stats.SuccessCount != 0 {
		t.Errorf("Expected SuccessCount 0, got %d", stats.SuccessCount)
	}
	if stats.SkipCount != 0 {
		t.Errorf("Expected SkipCount 0, got %d", stats.SkipCount)
	}
	if stats.ErrorCount != 0 {
		t.Errorf("Expected ErrorCount 0, got %d", stats.ErrorCount)
	}
	if stats.TotalMessages != 0 {
		t.Errorf("Expected TotalMessages 0, got %d", stats.TotalMessages)
	}
}

// Note: SyncProfiles のテストは Discord API のモックが必要なため、
// 統合テストで実装することを推奨します。
// ここではサービスの基本的な機能のみをテストしています。

func TestProfileService_SyncStats_ThreadSafe(t *testing.T) {
	profileRepo := newMockProfileRepository()
	userRepo := newMockUserRepository()

	service := NewProfileService(profileRepo, userRepo, "test-token", "test-channel")

	// 複数のゴルーチンから同時にstatsにアクセス
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			_ = service.GetLastSyncStats()
			done <- true
		}()
	}

	// すべてのゴルーチンが完了するまで待つ
	for i := 0; i < 10; i++ {
		<-done
	}

	// レースコンディションがなければテスト成功
	// go test -race で実行してレースコンディションを検出
}
