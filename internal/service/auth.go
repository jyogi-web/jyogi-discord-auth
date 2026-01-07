package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
	"github.com/jyogi-web/jyogi-discord-auth/internal/repository"
	"github.com/jyogi-web/jyogi-discord-auth/pkg/discord"
)

// AuthService は認証サービスを表します
type AuthService struct {
	discordClient *discord.Client
	userRepo      repository.UserRepository
	sessionRepo   repository.SessionRepository
	profileRepo   repository.ProfileRepository
	guildID       string
}

// NewAuthService は新しい認証サービスを作成します
func NewAuthService(
	discordClient *discord.Client,
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	profileRepo repository.ProfileRepository,
	guildID string,
) *AuthService {
	return &AuthService{
		discordClient: discordClient,
		userRepo:      userRepo,
		sessionRepo:   sessionRepo,
		profileRepo:   profileRepo,
		guildID:       guildID,
	}
}

// GenerateState はCSRF攻撃を防ぐためのランダムなstate文字列を生成します
func (s *AuthService) GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetAuthURL は認証URLを取得します
func (s *AuthService) GetAuthURL(state string) string {
	return s.discordClient.GetAuthURL(state)
}

// HandleCallback はDiscord OAuth2コールバックを処理します
// 認証成功時にセッショントークンを返します
func (s *AuthService) HandleCallback(ctx context.Context, code string) (string, error) {
	// 1. 認可コードをアクセストークンに交換
	token, err := s.discordClient.ExchangeCode(ctx, code)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code: %w", err)
	}

	// 2. ユーザー情報を取得
	discordUser, err := s.discordClient.GetUser(ctx, token)
	if err != nil {
		return "", fmt.Errorf("failed to get user info: %w", err)
	}

	// 3. じょぎサーバーのGuild Member情報を取得
	guildMember, err := s.discordClient.GetGuildMember(ctx, token, s.guildID)
	if err != nil {
		return "", fmt.Errorf("failed to get guild member: %w", err)
	}

	if guildMember == nil {
		return "", domain.ErrNotGuildMember
	}

	// 4. ユーザーをデータベースに保存または更新
	user, err := s.upsertUser(ctx, discordUser, guildMember)
	if err != nil {
		return "", fmt.Errorf("failed to upsert user: %w", err)
	}

	// 5. セッションを作成
	sessionToken, err := s.createSession(ctx, user.ID)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return sessionToken, nil
}

// upsertUser はユーザーを作成または更新します
func (s *AuthService) upsertUser(ctx context.Context, discordUser *discord.User, guildMember *discord.GuildMember) (*domain.User, error) {
	// 既存のユーザーを検索
	existingUser, err := s.userRepo.GetByDiscordID(ctx, discordUser.ID)
	if err != nil {
		// ユーザーが見つからない場合は新規作成
		if errors.Is(err, domain.ErrUserNotFound) {
			existingUser = nil
		} else {
			return nil, fmt.Errorf("failed to get user by discord_id: %w", err)
		}
	}

	now := time.Now()

	// GuildMember情報から追加フィールドを取得
	var guildNickname *string
	if guildMember.Nick != nil && *guildMember.Nick != "" {
		guildNickname = guildMember.Nick
	}

	var joinedAt *time.Time
	if guildMember.JoinedAt != "" {
		if parsedTime, err := time.Parse(time.RFC3339, guildMember.JoinedAt); err == nil {
			joinedAt = &parsedTime
		}
	}

	guildRoles := guildMember.Roles
	if guildRoles == nil || len(guildRoles) == 0 {
		guildRoles = []string{}
	}

	if existingUser != nil {
		// 既存ユーザーを更新
		existingUser.Username = discordUser.Username
		existingUser.DisplayName = discordUser.GetDisplayName()
		existingUser.AvatarURL = discordUser.GetAvatarURL()
		existingUser.GuildNickname = guildNickname
		existingUser.GuildRoles = guildRoles
		existingUser.JoinedAt = joinedAt
		existingUser.LastLoginAt = &now
		existingUser.UpdatedAt = now

		if err := s.userRepo.Update(ctx, existingUser); err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}

		return existingUser, nil
	}

	// 新規ユーザーを作成
	user := &domain.User{
		ID:            uuid.New().String(),
		DiscordID:     discordUser.ID,
		Username:      discordUser.Username,
		DisplayName:   discordUser.GetDisplayName(),
		AvatarURL:     discordUser.GetAvatarURL(),
		GuildNickname: guildNickname,
		GuildRoles:    guildRoles,
		JoinedAt:      joinedAt,
		CreatedAt:     now,
		UpdatedAt:     now,
		LastLoginAt:   &now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// createSession はセッションを作成します
func (s *AuthService) createSession(ctx context.Context, userID string) (string, error) {
	sessionToken, err := s.GenerateState() // ランダムなトークンを生成
	if err != nil {
		return "", fmt.Errorf("failed to generate session token: %w", err)
	}

	session := &domain.Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     sessionToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7日間有効
		CreatedAt: time.Now(),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return sessionToken, nil
}

// GetUserBySessionToken はセッショントークンからユーザーを取得します
func (s *AuthService) GetUserBySessionToken(ctx context.Context, sessionToken string) (*domain.User, error) {
	// セッションを取得
	session, err := s.sessionRepo.GetByToken(ctx, sessionToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// session is already checked for error above, if GetByToken returns error for not found,
	// err check handles it. If it returns valid session, session is not nil.
	// So we don't need explicit nil check if we trust GetByToken to return error on not found.
	// But let's keep it safe.
	if session == nil {
		return nil, fmt.Errorf("session not found")
	}

	// セッションが期限切れかチェック
	if session.IsExpired() {
		return nil, fmt.Errorf("session expired")
	}

	// ユーザーを取得
	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// Logout はセッションを削除してログアウトします
func (s *AuthService) Logout(ctx context.Context, sessionToken string) error {
	if err := s.sessionRepo.DeleteByToken(ctx, sessionToken); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// MemberWithProfile はメンバー情報とプロフィール情報を結合した構造体
type MemberWithProfile struct {
	User    *domain.User
	Profile *domain.Profile
}

// GetAllMembers は全てのメンバーを取得します（Deprecated: GetMembersWithProfilesを使用してください）
func (s *AuthService) GetAllMembers(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	// 全ユーザーを取得
	members, err := s.userRepo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get all members: %w", err)
	}

	return members, nil
}

// GetMembersWithProfiles は指定された範囲のメンバーとそのプロフィール情報を取得します
func (s *AuthService) GetMembersWithProfiles(ctx context.Context, limit, offset int) ([]*MemberWithProfile, error) {
	// ユーザーを取得
	users, err := s.userRepo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	if len(users) == 0 {
		return []*MemberWithProfile{}, nil
	}

	// ユーザーIDのリストを作成
	userIDs := make([]string, len(users))
	for i, user := range users {
		userIDs[i] = user.ID
	}

	// 該当するプロフィールを一括取得
	profiles, err := s.profileRepo.GetByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get profiles: %w", err)
	}

	// プロフィールをマップ化（UserID -> Profile）
	profileMap := make(map[string]*domain.Profile, len(profiles))
	for _, profile := range profiles {
		profileMap[profile.UserID] = profile
	}

	// ユーザーとプロフィールを結合
	result := make([]*MemberWithProfile, 0, len(users))
	for _, user := range users {
		result = append(result, &MemberWithProfile{
			User:    user,
			Profile: profileMap[user.ID], // マップから取得（存在しなければnil）
		})
	}

	return result, nil
}
