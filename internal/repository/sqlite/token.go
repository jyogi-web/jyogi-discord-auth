package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
	"github.com/jyogi-web/jyogi-discord-auth/internal/repository"
)

type tokenRepository struct {
	db *sql.DB
}

// NewTokenRepository は新しいSQLiteトークンリポジトリを作成します
func NewTokenRepository(db *sql.DB) repository.TokenRepository {
	return &tokenRepository{db: db}
}

// Create は新しいトークンをデータベースに挿入します
func (r *tokenRepository) Create(ctx context.Context, token *domain.Token) error {
	if err := token.Validate(); err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	query := `
		INSERT INTO tokens (id, token, token_type, user_id, client_id, expires_at, created_at, revoked)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		token.ID,
		token.Token,
		string(token.TokenType),
		token.UserID,
		token.ClientID,
		token.ExpiresAt.Format(time.RFC3339),
		token.CreatedAt.Format(time.RFC3339),
		token.Revoked,
	)
	if err != nil {
		return fmt.Errorf("failed to create token: %w", err)
	}

	return nil
}

// GetByToken はトークン文字列でトークンを取得します
func (r *tokenRepository) GetByToken(ctx context.Context, token string) (*domain.Token, error) {
	query := `
		SELECT id, token, token_type, user_id, client_id, expires_at, created_at, revoked
		FROM tokens
		WHERE token = ?
	`

	var expiresAtStr, createdAtStr string
	var tokenTypeStr string
	tokenObj := &domain.Token{}

	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&tokenObj.ID,
		&tokenObj.Token,
		&tokenTypeStr,
		&tokenObj.UserID,
		&tokenObj.ClientID,
		&expiresAtStr,
		&createdAtStr,
		&tokenObj.Revoked,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("token not found: %s", token)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	// TokenType を変換
	tokenObj.TokenType = domain.TokenType(tokenTypeStr)

	// 日付文字列をパース
	tokenObj.ExpiresAt, err = time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expires_at: %w", err)
	}
	tokenObj.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	return tokenObj, nil
}

// GetByUserID はユーザーのすべてのトークンを取得します
func (r *tokenRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Token, error) {
	query := `
		SELECT id, token, token_type, user_id, client_id, expires_at, created_at, revoked
		FROM tokens
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tokens by user_id: %w", err)
	}
	defer rows.Close()

	tokens := []*domain.Token{}
	for rows.Next() {
		var expiresAtStr, createdAtStr string
		var tokenTypeStr string
		token := &domain.Token{}

		if err := rows.Scan(
			&token.ID,
			&token.Token,
			&tokenTypeStr,
			&token.UserID,
			&token.ClientID,
			&expiresAtStr,
			&createdAtStr,
			&token.Revoked,
		); err != nil {
			return nil, fmt.Errorf("failed to scan token: %w", err)
		}

		// TokenType を変換
		token.TokenType = domain.TokenType(tokenTypeStr)

		// 日付文字列をパース
		token.ExpiresAt, err = time.Parse(time.RFC3339, expiresAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse expires_at: %w", err)
		}
		token.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}

		tokens = append(tokens, token)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tokens: %w", err)
	}

	return tokens, nil
}

// Revoke はトークンを取り消します
func (r *tokenRepository) Revoke(ctx context.Context, token string) error {
	query := `UPDATE tokens SET revoked = 1 WHERE token = ?`

	result, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("token not found: %s", token)
	}

	return nil
}

// DeleteExpired は期限切れのトークンをすべて削除します
func (r *tokenRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM tokens WHERE expires_at < ?`

	result, err := r.db.ExecContext(ctx, query, time.Now().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	// 削除されたトークン数をログ出力（オプション）
	if rowsAffected > 0 {
		fmt.Printf("Deleted %d expired tokens\n", rowsAffected)
	}

	return nil
}
