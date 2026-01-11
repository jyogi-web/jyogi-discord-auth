package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
)

// モックTokenRepository
type mockTokenRepository struct {
	tokens        map[string]*domain.Token
	getByTokenErr error
	createError   error
	revokeError   error
}

func newMockTokenRepository() *mockTokenRepository {
	return &mockTokenRepository{
		tokens: make(map[string]*domain.Token),
	}
}

func (m *mockTokenRepository) Create(ctx context.Context, token *domain.Token) error {
	if m.createError != nil {
		return m.createError
	}
	m.tokens[token.Token] = token
	return nil
}

func (m *mockTokenRepository) GetByToken(ctx context.Context, token string) (*domain.Token, error) {
	if m.getByTokenErr != nil {
		return nil, m.getByTokenErr
	}
	t, ok := m.tokens[token]
	if !ok {
		return nil, errors.New("token not found")
	}
	return t, nil
}

func (m *mockTokenRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Token, error) {
	var tokens []*domain.Token
	for _, t := range m.tokens {
		if t.UserID == userID {
			tokens = append(tokens, t)
		}
	}
	return tokens, nil
}

func (m *mockTokenRepository) Revoke(ctx context.Context, token string) error {
	if m.revokeError != nil {
		return m.revokeError
	}
	if t, ok := m.tokens[token]; ok {
		t.Revoked = true
	}
	return nil
}

func (m *mockTokenRepository) DeleteExpired(ctx context.Context) error {
	return nil
}

// モックClientRepository
type mockClientRepository struct {
	clients map[string]*domain.ClientApp
}

func newMockClientRepository() *mockClientRepository {
	return &mockClientRepository{
		clients: make(map[string]*domain.ClientApp),
	}
}

func (m *mockClientRepository) Create(ctx context.Context, client *domain.ClientApp) error {
	m.clients[client.ClientID] = client
	return nil
}

func (m *mockClientRepository) GetByID(ctx context.Context, id string) (*domain.ClientApp, error) {
	for _, c := range m.clients {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, errors.New("client not found")
}

func (m *mockClientRepository) GetByClientID(ctx context.Context, clientID string) (*domain.ClientApp, error) {
	c, ok := m.clients[clientID]
	if !ok {
		return nil, errors.New("client not found")
	}
	return c, nil
}

func (m *mockClientRepository) GetAll(ctx context.Context) ([]*domain.ClientApp, error) {
	var clients []*domain.ClientApp
	for _, c := range m.clients {
		clients = append(clients, c)
	}
	return clients, nil
}

func (m *mockClientRepository) Update(ctx context.Context, client *domain.ClientApp) error {
	m.clients[client.ClientID] = client
	return nil
}

func (m *mockClientRepository) Delete(ctx context.Context, id string) error {
	for clientID, c := range m.clients {
		if c.ID == id {
			delete(m.clients, clientID)
			return nil
		}
	}
	return nil
}

func (m *mockClientRepository) ValidateRedirectURI(ctx context.Context, clientID, redirectURI string) (bool, error) {
	return true, nil
}

// モックAuthCodeRepository
type mockAuthCodeRepository struct {
	authCodes map[string]*domain.AuthCode
}

func newMockAuthCodeRepository() *mockAuthCodeRepository {
	return &mockAuthCodeRepository{
		authCodes: make(map[string]*domain.AuthCode),
	}
}

func (m *mockAuthCodeRepository) Create(ctx context.Context, authCode *domain.AuthCode) error {
	m.authCodes[authCode.Code] = authCode
	return nil
}

func (m *mockAuthCodeRepository) GetByCode(ctx context.Context, code string) (*domain.AuthCode, error) {
	ac, ok := m.authCodes[code]
	if !ok {
		return nil, errors.New("auth code not found")
	}
	return ac, nil
}

func (m *mockAuthCodeRepository) MarkAsUsed(ctx context.Context, code string) error {
	if ac, ok := m.authCodes[code]; ok {
		ac.Used = true
	}
	return nil
}

func (m *mockAuthCodeRepository) DeleteExpired(ctx context.Context) error {
	return nil
}

// TestOAuth2Service_GetUserByAccessToken_Success tests that a valid access token returns the expected user
func TestOAuth2Service_GetUserByAccessToken_Success(t *testing.T) {
	tokenRepo := newMockTokenRepository()
	userRepo := newMockUserRepository()
	clientRepo := newMockClientRepository()
	authCodeRepo := newMockAuthCodeRepository()

	service := NewOAuth2Service(clientRepo, authCodeRepo, tokenRepo, userRepo)

	// テストデータを準備
	userID := uuid.New().String()
	tokenString := "valid-access-token"

	expectedUser := &domain.User{
		ID:          userID,
		DiscordID:   "123456789",
		Username:    "testuser",
		DisplayName: "Test User",
		AvatarURL:   "https://example.com/avatar.png",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	validToken := &domain.Token{
		ID:        uuid.New().String(),
		Token:     tokenString,
		TokenType: domain.TokenTypeAccess,
		UserID:    userID,
		ClientID:  "test-client",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
		Revoked:   false,
	}

	tokenRepo.tokens[tokenString] = validToken
	userRepo.users[userID] = expectedUser

	ctx := context.Background()
	user, err := service.GetUserByAccessToken(ctx, tokenString)

	if err != nil {
		t.Fatalf("GetUserByAccessToken failed: %v", err)
	}

	if user == nil {
		t.Fatal("Expected user, got nil")
	}

	if user.ID != expectedUser.ID {
		t.Errorf("User ID mismatch: got %v, want %v", user.ID, expectedUser.ID)
	}

	if user.Username != expectedUser.Username {
		t.Errorf("Username mismatch: got %v, want %v", user.Username, expectedUser.Username)
	}
}

// TestOAuth2Service_GetUserByAccessToken_TokenNotFound tests that token not found returns an error
func TestOAuth2Service_GetUserByAccessToken_TokenNotFound(t *testing.T) {
	tokenRepo := newMockTokenRepository()
	userRepo := newMockUserRepository()
	clientRepo := newMockClientRepository()
	authCodeRepo := newMockAuthCodeRepository()

	service := NewOAuth2Service(clientRepo, authCodeRepo, tokenRepo, userRepo)

	ctx := context.Background()
	_, err := service.GetUserByAccessToken(ctx, "non-existent-token")

	if err == nil {
		t.Fatal("Expected error for non-existent token, got nil")
	}
}

// TestOAuth2Service_GetUserByAccessToken_WrongTokenType tests that a refresh token returns an error
func TestOAuth2Service_GetUserByAccessToken_WrongTokenType(t *testing.T) {
	tokenRepo := newMockTokenRepository()
	userRepo := newMockUserRepository()
	clientRepo := newMockClientRepository()
	authCodeRepo := newMockAuthCodeRepository()

	service := NewOAuth2Service(clientRepo, authCodeRepo, tokenRepo, userRepo)

	userID := uuid.New().String()
	tokenString := "refresh-token"

	// リフレッシュトークンを作成
	refreshToken := &domain.Token{
		ID:        uuid.New().String(),
		Token:     tokenString,
		TokenType: domain.TokenTypeRefresh, // Not an access token
		UserID:    userID,
		ClientID:  "test-client",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now(),
		Revoked:   false,
	}

	tokenRepo.tokens[tokenString] = refreshToken

	ctx := context.Background()
	_, err := service.GetUserByAccessToken(ctx, tokenString)

	if err == nil {
		t.Fatal("Expected error for refresh token, got nil")
	}

	expectedErrMsg := "token is not an access token"
	if err.Error() != expectedErrMsg {
		t.Errorf("Error message mismatch: got %v, want %v", err.Error(), expectedErrMsg)
	}
}

// TestOAuth2Service_GetUserByAccessToken_ExpiredToken tests that an expired token returns an error
func TestOAuth2Service_GetUserByAccessToken_ExpiredToken(t *testing.T) {
	tokenRepo := newMockTokenRepository()
	userRepo := newMockUserRepository()
	clientRepo := newMockClientRepository()
	authCodeRepo := newMockAuthCodeRepository()

	service := NewOAuth2Service(clientRepo, authCodeRepo, tokenRepo, userRepo)

	userID := uuid.New().String()
	tokenString := "expired-token"

	// 期限切れのトークンを作成
	expiredToken := &domain.Token{
		ID:        uuid.New().String(),
		Token:     tokenString,
		TokenType: domain.TokenTypeAccess,
		UserID:    userID,
		ClientID:  "test-client",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // 1時間前に期限切れ
		CreatedAt: time.Now().Add(-2 * time.Hour),
		Revoked:   false,
	}

	tokenRepo.tokens[tokenString] = expiredToken

	ctx := context.Background()
	_, err := service.GetUserByAccessToken(ctx, tokenString)

	if err == nil {
		t.Fatal("Expected error for expired token, got nil")
	}

	expectedErrMsg := "token is expired or revoked"
	if err.Error() != expectedErrMsg {
		t.Errorf("Error message mismatch: got %v, want %v", err.Error(), expectedErrMsg)
	}
}

// TestOAuth2Service_GetUserByAccessToken_RevokedToken tests that a revoked token returns an error
func TestOAuth2Service_GetUserByAccessToken_RevokedToken(t *testing.T) {
	tokenRepo := newMockTokenRepository()
	userRepo := newMockUserRepository()
	clientRepo := newMockClientRepository()
	authCodeRepo := newMockAuthCodeRepository()

	service := NewOAuth2Service(clientRepo, authCodeRepo, tokenRepo, userRepo)

	userID := uuid.New().String()
	tokenString := "revoked-token"

	// 取り消し済みのトークンを作成
	revokedToken := &domain.Token{
		ID:        uuid.New().String(),
		Token:     tokenString,
		TokenType: domain.TokenTypeAccess,
		UserID:    userID,
		ClientID:  "test-client",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
		Revoked:   true, // Revoked
	}

	tokenRepo.tokens[tokenString] = revokedToken

	ctx := context.Background()
	_, err := service.GetUserByAccessToken(ctx, tokenString)

	if err == nil {
		t.Fatal("Expected error for revoked token, got nil")
	}

	expectedErrMsg := "token is expired or revoked"
	if err.Error() != expectedErrMsg {
		t.Errorf("Error message mismatch: got %v, want %v", err.Error(), expectedErrMsg)
	}
}

// TestOAuth2Service_GetUserByAccessToken_UserNotFound tests that userRepo.GetByID error is propagated
func TestOAuth2Service_GetUserByAccessToken_UserNotFound(t *testing.T) {
	tokenRepo := newMockTokenRepository()
	userRepo := newMockUserRepository()
	clientRepo := newMockClientRepository()
	authCodeRepo := newMockAuthCodeRepository()

	service := NewOAuth2Service(clientRepo, authCodeRepo, tokenRepo, userRepo)

	userID := uuid.New().String()
	tokenString := "valid-token-but-user-not-found"

	validToken := &domain.Token{
		ID:        uuid.New().String(),
		Token:     tokenString,
		TokenType: domain.TokenTypeAccess,
		UserID:    userID, // このユーザーは存在しない
		ClientID:  "test-client",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
		Revoked:   false,
	}

	tokenRepo.tokens[tokenString] = validToken
	// userRepoにはユーザーを追加しない

	ctx := context.Background()
	_, err := service.GetUserByAccessToken(ctx, tokenString)

	if err == nil {
		t.Fatal("Expected error when user not found, got nil")
	}

	// エラーが "user not found" を含むことを確認
	if err.Error() != "failed to get user: user not found" {
		t.Errorf("Error message mismatch: got %v", err.Error())
	}
}
