package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
	"github.com/jyogi-web/jyogi-discord-auth/internal/repository"
	"github.com/jyogi-web/jyogi-discord-auth/pkg/auth"
)

const (
	// AuthCodeExpiration は認可コードの有効期限（10分）
	AuthCodeExpiration = 10 * time.Minute
	// AccessTokenExpiration はアクセストークンの有効期限（1時間）
	AccessTokenExpiration = 1 * time.Hour
	// RefreshTokenExpiration はリフレッシュトークンの有効期限（7日）
	RefreshTokenExpiration = 7 * 24 * time.Hour
)

// OAuth2Service はOAuth2認可サーバーの機能を提供します
type OAuth2Service struct {
	clientRepo   repository.ClientRepository
	authCodeRepo repository.AuthCodeRepository
	tokenRepo    repository.TokenRepository
	userRepo     repository.UserRepository
}

// NewOAuth2Service は新しいOAuth2サービスを作成します
func NewOAuth2Service(
	clientRepo repository.ClientRepository,
	authCodeRepo repository.AuthCodeRepository,
	tokenRepo repository.TokenRepository,
	userRepo repository.UserRepository,
) *OAuth2Service {
	return &OAuth2Service{
		clientRepo:   clientRepo,
		authCodeRepo: authCodeRepo,
		tokenRepo:    tokenRepo,
		userRepo:     userRepo,
	}
}

// AuthorizeRequest は認可リクエストのパラメータを表します
type AuthorizeRequest struct {
	ClientID     string
	RedirectURI  string
	ResponseType string
	State        string
	UserID       string // セッションから取得したユーザーID
}

// AuthorizeResponse は認可レスポンスを表します
type AuthorizeResponse struct {
	Code        string
	State       string
	RedirectURI string
}

// Authorize はOAuth2認可リクエストを処理し、認可コードを発行します
func (s *OAuth2Service) Authorize(ctx context.Context, req *AuthorizeRequest) (*AuthorizeResponse, error) {
	// 1. response_type の検証
	if req.ResponseType != "code" {
		return nil, fmt.Errorf("unsupported response_type: %s", req.ResponseType)
	}

	// 2. クライアントの検証
	client, err := s.clientRepo.GetByClientID(ctx, req.ClientID)
	if err != nil {
		return nil, fmt.Errorf("invalid client_id: %w", err)
	}

	// 3. redirect_uri の検証
	valid, err := s.clientRepo.ValidateRedirectURI(ctx, req.ClientID, req.RedirectURI)
	if err != nil {
		return nil, fmt.Errorf("failed to validate redirect_uri: %w", err)
	}
	if !valid {
		return nil, fmt.Errorf("invalid redirect_uri: %s", req.RedirectURI)
	}

	// 4. ユーザーの検証
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}

	// 5. 認可コードを生成
	code, err := generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth code: %w", err)
	}

	// 6. 認可コードを保存
	authCode := &domain.AuthCode{
		ID:          uuid.New().String(),
		Code:        code,
		ClientID:    client.ClientID,
		UserID:      user.ID,
		RedirectURI: req.RedirectURI,
		ExpiresAt:   time.Now().Add(AuthCodeExpiration),
		CreatedAt:   time.Now(),
		Used:        false,
	}

	if err := s.authCodeRepo.Create(ctx, authCode); err != nil {
		return nil, fmt.Errorf("failed to store auth code: %w", err)
	}

	return &AuthorizeResponse{
		Code:        code,
		State:       req.State,
		RedirectURI: req.RedirectURI,
	}, nil
}

// TokenRequest はトークンリクエストのパラメータを表します
type TokenRequest struct {
	GrantType    string
	Code         string
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// TokenResponse はトークンレスポンスを表します
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// ExchangeToken は認可コードをアクセストークンに交換します
func (s *OAuth2Service) ExchangeToken(ctx context.Context, req *TokenRequest) (*TokenResponse, error) {
	// 1. grant_type の検証
	if req.GrantType != "authorization_code" {
		return nil, fmt.Errorf("unsupported grant_type: %s", req.GrantType)
	}

	// 2. クライアント認証
	client, err := s.clientRepo.GetByClientID(ctx, req.ClientID)
	if err != nil {
		return nil, fmt.Errorf("invalid client_id: %w", err)
	}

	// クライアントシークレットの検証
	if err := auth.ValidateClientSecret(req.ClientSecret, client.ClientSecret); err != nil {
		return nil, fmt.Errorf("invalid client_secret: %w", err)
	}

	// 3. 認可コードの取得と検証
	authCode, err := s.authCodeRepo.GetByCode(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("invalid authorization code: %w", err)
	}

	// 認可コードが既に使用されていないか確認
	if authCode.Used {
		return nil, fmt.Errorf("authorization code already used")
	}

	// 認可コードの有効期限確認
	if authCode.IsExpired() {
		return nil, fmt.Errorf("authorization code expired")
	}

	// クライアントIDの一致確認
	if authCode.ClientID != client.ClientID {
		return nil, fmt.Errorf("client_id mismatch")
	}

	// redirect_uri の一致確認
	if authCode.RedirectURI != req.RedirectURI {
		return nil, fmt.Errorf("redirect_uri mismatch")
	}

	// 4. 認可コードを使用済みにマーク
	if err := s.authCodeRepo.MarkAsUsed(ctx, req.Code); err != nil {
		return nil, fmt.Errorf("failed to mark auth code as used: %w", err)
	}

	// 5. アクセストークンとリフレッシュトークンを生成
	accessToken, err := generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	now := time.Now()

	// 6. アクセストークンを保存
	accessTokenObj := &domain.Token{
		ID:        uuid.New().String(),
		Token:     accessToken,
		TokenType: domain.TokenTypeAccess,
		UserID:    authCode.UserID,
		ClientID:  client.ClientID,
		ExpiresAt: now.Add(AccessTokenExpiration),
		CreatedAt: now,
		Revoked:   false,
	}
	if err := s.tokenRepo.Create(ctx, accessTokenObj); err != nil {
		return nil, fmt.Errorf("failed to store access token: %w", err)
	}

	// 7. リフレッシュトークンを保存
	refreshTokenObj := &domain.Token{
		ID:        uuid.New().String(),
		Token:     refreshToken,
		TokenType: domain.TokenTypeRefresh,
		UserID:    authCode.UserID,
		ClientID:  client.ClientID,
		ExpiresAt: now.Add(RefreshTokenExpiration),
		CreatedAt: now,
		Revoked:   false,
	}
	if err := s.tokenRepo.Create(ctx, refreshTokenObj); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(AccessTokenExpiration.Seconds()),
		RefreshToken: refreshToken,
	}, nil
}

// generateSecureToken は暗号学的に安全なランダムトークンを生成します
func generateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
