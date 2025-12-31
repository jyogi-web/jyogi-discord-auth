package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
	"github.com/jyogi-web/jyogi-discord-auth/internal/repository"
)

type clientRepository struct {
	db *sql.DB
}

// NewClientRepository は新しいSQLiteクライアントリポジトリを作成します
func NewClientRepository(db *sql.DB) repository.ClientRepository {
	return &clientRepository{db: db}
}

// Create は新しいクライアントアプリをデータベースに挿入します
func (r *clientRepository) Create(ctx context.Context, client *domain.ClientApp) error {
	if err := client.Validate(); err != nil {
		return fmt.Errorf("invalid client: %w", err)
	}

	// RedirectURIsをJSON文字列に変換
	redirectURIsJSON, err := json.Marshal(client.RedirectURIs)
	if err != nil {
		return fmt.Errorf("failed to marshal redirect_uris: %w", err)
	}

	query := `
		INSERT INTO client_apps (id, client_id, name, client_secret, redirect_uris, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecContext(ctx, query,
		client.ID,
		client.ClientID,
		client.Name,
		client.ClientSecret,
		string(redirectURIsJSON),
		client.CreatedAt.Format(time.RFC3339),
		client.UpdatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	return nil
}

// GetByID はIDでクライアントアプリを取得します
func (r *clientRepository) GetByID(ctx context.Context, id string) (*domain.ClientApp, error) {
	query := `
		SELECT id, client_id, name, client_secret, redirect_uris, created_at, updated_at
		FROM client_apps
		WHERE id = ?
	`

	var redirectURIsJSON string
	var createdAtStr, updatedAtStr string
	client := &domain.ClientApp{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&client.ID,
		&client.ClientID,
		&client.Name,
		&client.ClientSecret,
		&redirectURIsJSON,
		&createdAtStr,
		&updatedAtStr,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("client not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	// RedirectURIsをデシリアライズ
	if err := json.Unmarshal([]byte(redirectURIsJSON), &client.RedirectURIs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal redirect_uris: %w", err)
	}

	// 日付文字列をパース
	client.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}
	client.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	return client, nil
}

// GetByClientID はクライアントIDでクライアントアプリを取得します
func (r *clientRepository) GetByClientID(ctx context.Context, clientID string) (*domain.ClientApp, error) {
	query := `
		SELECT id, client_id, name, client_secret, redirect_uris, created_at, updated_at
		FROM client_apps
		WHERE client_id = ?
	`

	var redirectURIsJSON string
	var createdAtStr, updatedAtStr string
	client := &domain.ClientApp{}

	err := r.db.QueryRowContext(ctx, query, clientID).Scan(
		&client.ID,
		&client.ClientID,
		&client.Name,
		&client.ClientSecret,
		&redirectURIsJSON,
		&createdAtStr,
		&updatedAtStr,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("client not found: %s", clientID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	// RedirectURIsをデシリアライズ
	if err := json.Unmarshal([]byte(redirectURIsJSON), &client.RedirectURIs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal redirect_uris: %w", err)
	}

	// 日付文字列をパース
	client.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}
	client.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	return client, nil
}

// GetAll はすべてのクライアントアプリを取得します
func (r *clientRepository) GetAll(ctx context.Context) ([]*domain.ClientApp, error) {
	query := `
		SELECT id, client_id, name, client_secret, redirect_uris, created_at, updated_at
		FROM client_apps
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all clients: %w", err)
	}
	defer rows.Close()

	clients := []*domain.ClientApp{}
	for rows.Next() {
		var redirectURIsJSON string
		var createdAtStr, updatedAtStr string
		client := &domain.ClientApp{}

		if err := rows.Scan(
			&client.ID,
			&client.ClientID,
			&client.Name,
			&client.ClientSecret,
			&redirectURIsJSON,
			&createdAtStr,
			&updatedAtStr,
		); err != nil {
			return nil, fmt.Errorf("failed to scan client: %w", err)
		}

		// RedirectURIsをデシリアライズ
		if err := json.Unmarshal([]byte(redirectURIsJSON), &client.RedirectURIs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal redirect_uris: %w", err)
		}

		// 日付文字列をパース
		client.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}
		client.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse updated_at: %w", err)
		}

		clients = append(clients, client)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating clients: %w", err)
	}

	return clients, nil
}

// Update はクライアントアプリを更新します
func (r *clientRepository) Update(ctx context.Context, client *domain.ClientApp) error {
	if err := client.Validate(); err != nil {
		return fmt.Errorf("invalid client: %w", err)
	}

	// RedirectURIsをJSON文字列に変換
	redirectURIsJSON, err := json.Marshal(client.RedirectURIs)
	if err != nil {
		return fmt.Errorf("failed to marshal redirect_uris: %w", err)
	}

	query := `
		UPDATE client_apps
		SET client_id = ?, name = ?, client_secret = ?, redirect_uris = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		client.ClientID,
		client.Name,
		client.ClientSecret,
		string(redirectURIsJSON),
		time.Now().Format(time.RFC3339),
		client.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update client: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("client not found: %s", client.ID)
	}

	return nil
}

// Delete はIDでクライアントアプリを削除します
func (r *clientRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM client_apps WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("client not found: %s", id)
	}

	return nil
}

// ValidateRedirectURI はリダイレクトURIがクライアントアプリに登録されているか検証します
func (r *clientRepository) ValidateRedirectURI(ctx context.Context, clientID, redirectURI string) (bool, error) {
	client, err := r.GetByClientID(ctx, clientID)
	if err != nil {
		return false, err
	}

	// リダイレクトURIのリストに含まれているか確認
	for _, uri := range client.RedirectURIs {
		if uri == redirectURI {
			return true, nil
		}
	}

	return false, nil
}
