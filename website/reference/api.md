# API リファレンス

## 認証 (Auth)

### Discordログイン

- **URL**: `/auth/login`
- **Method**: `GET`
- **Description**: DiscordのOAuth2認証フローを開始します。

### Discordコールバック

- **URL**: `/auth/callback`
- **Method**: `GET`
- **Description**: Discordからのリダイレクトを受け取り、セッションを作成します。

### ログアウト

- **URL**: `/auth/logout`
- **Method**: `POST`
- **Description**: セッションを破棄してログアウトします。

## トークン (Token)

### JWT発行

- **URL**: `/token`
- **Method**: `POST`
- **Header**: `Cookie: session=...`
- **Description**: 有効なセッションを持つユーザーに対してJWTを発行します。

### トークン更新

- **URL**: `/token/refresh`
- **Method**: `POST`
- **Description**: リフレッシュトークンを使用して新しいアクセストークンを取得します。

## 保護されたリソース (Protected)

以下のエンドポイントは有効なJWTが必要です。

### JWT検証

- **URL**: `/api/verify`
- **Method**: `GET`
- **Header**: `Authorization: Bearer <token>`
- **Description**: トークンの有効性を検証します。

### ユーザー情報取得

- **URL**: `/api/user`
- **Method**: `GET`
- **Header**: `Authorization: Bearer <token>`
- **Description**: ログイン中のユーザー情報を返します。
