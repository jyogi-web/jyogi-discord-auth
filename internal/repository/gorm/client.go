package gorm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
	"github.com/jyogi-web/jyogi-discord-auth/internal/repository"
)

type clientRepository struct {
	db *gorm.DB
}

// NewClientRepository は新しいGORMクライアントアプリリポジトリを作成します
func NewClientRepository(db *gorm.DB) repository.ClientRepository {
	return &clientRepository{db: db}
}

// Create は新しいクライアントアプリをデータベースに挿入します
func (r *clientRepository) Create(ctx context.Context, client *domain.ClientApp) error {
	if err := client.Validate(); err != nil {
		return fmt.Errorf("invalid client: %w", err)
	}

	c, err := FromDomainClientApp(client)
	if err != nil {
		return fmt.Errorf("failed to map client app: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(c).Error; err != nil {
		return fmt.Errorf("failed to create client app: %w", err)
	}

	return nil
}

// GetByID はIDでクライアントアプリを取得します
func (r *clientRepository) GetByID(ctx context.Context, id string) (*domain.ClientApp, error) {
	var c ClientApp
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("client app not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get client app: %w", err)
	}

	return c.ToDomain()
}

// GetByClientID はClientIDでクライアントアプリを取得します
func (r *clientRepository) GetByClientID(ctx context.Context, clientID string) (*domain.ClientApp, error) {
	var c ClientApp
	if err := r.db.WithContext(ctx).Where("client_id = ?", clientID).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("client app not found: %s", clientID)
		}
		return nil, fmt.Errorf("failed to get client app: %w", err)
	}

	return c.ToDomain()
}

// GetAll はすべてのクライアントアプリを取得します
func (r *clientRepository) GetAll(ctx context.Context) ([]*domain.ClientApp, error) {
	var clients []ClientApp
	if err := r.db.WithContext(ctx).Find(&clients).Error; err != nil {
		return nil, fmt.Errorf("failed to get client apps: %w", err)
	}

	domainClients := make([]*domain.ClientApp, len(clients))
	for i, c := range clients {
		dc, err := c.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("failed to map client app (id: %s): %w", c.ID, err)
		}
		domainClients[i] = dc
	}

	return domainClients, nil
}

// Update は既存のクライアントアプリを更新します
func (r *clientRepository) Update(ctx context.Context, client *domain.ClientApp) error {
	if err := client.Validate(); err != nil {
		return fmt.Errorf("invalid client: %w", err)
	}

	c, err := FromDomainClientApp(client)
	if err != nil {
		return fmt.Errorf("failed to map client app: %w", err)
	}

	c.UpdatedAt = time.Now()

	// 全フィールド更新。IDで特定
	result := r.db.WithContext(ctx).Model(&ClientApp{}).Where("id = ?", c.ID).Updates(map[string]interface{}{
		"name":          c.Name,
		"redirect_uris": c.RedirectURIs,
		"updated_at":    c.UpdatedAt,
	})

	if result.Error != nil {
		return fmt.Errorf("failed to update client app: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("client app not found: %s", client.ID)
	}

	return nil
}

// Delete はクライアントアプリをデータベースから削除します
func (r *clientRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&ClientApp{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete client app: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("client app not found: %s", id)
	}

	return nil
}

// ValidateRedirectURI は指定されたURIがクライアントに許可されているか確認します
func (r *clientRepository) ValidateRedirectURI(ctx context.Context, clientID, redirectURI string) (bool, error) {
	client, err := r.GetByClientID(ctx, clientID)
	if err != nil {
		return false, err
	}

	return client.IsRedirectURIValid(redirectURI), nil
}
