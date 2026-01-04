# じょぎメンバー認証システム

> [!NOTE]
> このリポジトリは GitHub Pages や外部ツール連携（CodeRabbit 等）の利便性のために公開設定としていますが、**「じょぎ」サークル内部での利用を目的とした認証基盤**です。外部ユーザーへのサポートや、汎用的な利用は想定しておりませんのでご了承ください。

Discord OAuth2を使用したじょぎメンバー専用の認証システム。Discordアカウントでログインし、じょぎサーバーのメンバーシップを確認後、JWTを発行する。他の内製ツールがSSOとして利用できる認証基盤を提供します。

## 📚 ドキュメント

詳細なドキュメントは公式サイトをご覧ください：

**🌐 [https://TODO](https://TODO)** （日本語 / English）

- **クイックスタート（クライアント統合）**: 他のアプリから認証システムを利用する方法
- **開発者ガイド**: 開発環境のセットアップと貢献方法
- **アーキテクチャ**: システム設計とプロジェクト構造
- **デプロイメント**: Cloud RunやCloud Functionsへのデプロイ手順
- **APIリファレンス**: 利用可能なエンドポイントの詳細

## 概要

じょぎ内製ツール作成にあたり基盤となる認証システムです。Discordがメインのチャットツールであるため、DiscordアカウントをIdPとして活用し、SSOを実現します。

### 主な機能

- **Discord OAuth2ログイン**: Discordアカウントでログイン
- **メンバーシップ確認**: じょぎサーバーメンバーのみアクセス許可
- **JWT発行**: 認証成功後、アクセストークン（JWT）を発行
- **SSO対応**: 他の内製ツールがこの認証サーバーを利用可能
- **セッション管理**: ユーザーのログイン状態を管理
- **プロフィール同期**: Discord自己紹介チャンネルから自動的にメンバープロフィールを取得・更新

## クイックスタート

### 開発者向け

開発環境のセットアップについては、[開発者ガイド](https://TODO)をご覧ください。

```bash
# リポジトリのクローン
git clone https://github.com/jyogi-web/jyogi-discord-auth.git
cd jyogi-discord-auth

# 環境変数の設定
cp .env.example .env
# .envファイルを編集

# サーバー起動
go run cmd/server/main.go
```

### クライアント統合者向け

他のアプリから認証システムを利用する方法については、[クイックスタート（クライアント統合）](https://TODO)をご覧ください。

## API エンドポイント

詳細なAPI仕様は[APIリファレンス](https://TODO)をご覧ください。

### 認証

- `GET /auth/login` - Discordログイン
- `GET /auth/callback` - Discordコールバック
- `POST /auth/logout` - ログアウト

### OAuth2 (SSO)

- `GET /oauth/authorize` - 認可リクエスト
- `POST /oauth/token` - トークン取得

### API (Protected)

- `GET /api/verify` - JWT検証
- `GET /api/user` - ユーザー情報取得

## 貢献

プルリクエストを歓迎します。大きな変更の場合は、まずissueを開いて変更内容を議論してください。

## サポート

問題が発生した場合は、[GitHub Issues](https://github.com/jyogi-web/jyogi-discord-auth/issues)で報告してください。
