package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
	"github.com/jyogi-web/jyogi-discord-auth/internal/repository"
	"github.com/jyogi-web/jyogi-discord-auth/pkg/auth"
)

// ClientService はクライアントアプリケーションの管理機能を提供します
type ClientService struct {
	clientRepo repository.ClientRepository
}

// NewClientService は新しいClientServiceを作成します
func NewClientService(clientRepo repository.ClientRepository) *ClientService {
	return &ClientService{
		clientRepo: clientRepo,
	}
}

// RegisterClient は新しいクライアントアプリケーションを登録します
// clientSecret は平文で受け取り、内部でハッシュ化して保存します
func (s *ClientService) RegisterClient(ctx context.Context, ownerID, clientID, plainSecret, name string, redirectURIs []string) (*domain.ClientApp, error) {
	if ownerID == "" {
		return nil, fmt.Errorf("owner_id is required")
	}
	if clientID == "" {
		return nil, fmt.Errorf("client_id is required")
	}
	if plainSecret == "" {
		return nil, fmt.Errorf("client_secret is required")
	}
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if len(redirectURIs) == 0 {
		return nil, fmt.Errorf("at least one redirect_uri is required")
	}

	// クライアントIDの重複チェック
	_, err := s.clientRepo.GetByClientID(ctx, clientID)
	if err == nil {
		return nil, fmt.Errorf("client_id already exists: %s", clientID)
	}

	// シークレットのハッシュ化
	hashedSecret, err := auth.HashClientSecret(plainSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to hash client secret: %w", err)
	}

	now := time.Now()
	client := &domain.ClientApp{
		ID:           uuid.New().String(),
		OwnerID:      ownerID,
		ClientID:     clientID,
		ClientSecret: hashedSecret,
		Name:         name,
		RedirectURIs: redirectURIs,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.clientRepo.Create(ctx, client); err != nil {
		return nil, fmt.Errorf("failed to create client app: %w", err)
	}

	return client, nil
}

// UpdateClient は既存のクライアントアプリケーションを更新します
func (s *ClientService) UpdateClient(ctx context.Context, clientID, plainSecret, name string, redirectURIs []string) (*domain.ClientApp, error) {
	// クライアントの取得
	client, err := s.clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("client not found: %w", err)
	}

	// フィールドの更新
	if name != "" {
		client.Name = name
	}
	if len(redirectURIs) > 0 {
		client.RedirectURIs = redirectURIs
	}

	// シークレットが指定されている場合のみ更新
	if plainSecret != "" {
		hashedSecret, err := auth.HashClientSecret(plainSecret)
		if err != nil {
			return nil, fmt.Errorf("failed to hash client secret: %w", err)
		}
		client.ClientSecret = hashedSecret
	}

	client.UpdatedAt = time.Now()

	if err := s.clientRepo.Update(ctx, client); err != nil {
		return nil, fmt.Errorf("failed to update client app: %w", err)
	}

	return client, nil
}

// GetAllClients は全てのクライアントアプリケーションを取得します
func (s *ClientService) GetAllClients(ctx context.Context) ([]*domain.ClientApp, error) {
	clients, err := s.clientRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all clients: %w", err)
	}
	return clients, nil
}

// GetClientByID はIDからクライアントアプリケーションを取得します
func (s *ClientService) GetClientByID(ctx context.Context, id string) (*domain.ClientApp, error) {
	client, err := s.clientRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}
	return client, nil
}

// DeleteClient はクライアントアプリケーションを削除します
func (s *ClientService) DeleteClient(ctx context.Context, id string) error {
	if err := s.clientRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}
	return nil
}
