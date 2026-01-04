# クイックスタート (クライアント統合)

Next.jsプロジェクトをゼロから作成し、「じょぎメンバー認証」を統合するまでの完全なハンズオンガイドです。
App Router を使用して、セキュアな認証フローを実装します。

## 概要

このガイドで作成するもの：

1. Next.jsアプリケーション
2. ログインボタン (認証サーバーへリダイレクト)
3. コールバックハンドラ (認可コードをトークンに交換)
4. プロフィール表示ページ (取得したトークンでAPIアクセス)

## Step 1: Next.jsプロジェクトの作成

まずは新しいNext.jsプロジェクトを作成します。

```bash
npx create-next-app@latest jyogi-client-demo
# 設定はすべてデフォルト(Enter)でOKです
cd jyogi-client-demo
```

## Step 2: 環境変数の設定

プロジェクトのルートに `.env.local` ファイルを作成し、認証情報を設定します。
これらの値は認証サーバーの管理者から取得してください。

```bash
# .env.local
NEXT_PUBLIC_AUTH_SERVER_URL="https://your-auth-server-url.com"
CLIENT_ID="your_client_id"
CLIENT_SECRET="your_client_secret"
REDIRECT_URI="http://localhost:3000/api/auth/callback"
```

## Step 3: ログインボタンの作成

トップページにログインリンクを追加します。
CSRF対策のために `state` パラメータを生成するのがベストプラクティスですが、ここでは簡易化のために固定値を使用します。

`app/page.tsx`:

```tsx
import Link from 'next/link'

export default function Home() {
  const authUrl = `${process.env.NEXT_PUBLIC_AUTH_SERVER_URL}/oauth/authorize` +
    `?client_id=${process.env.CLIENT_ID}` +
    `&redirect_uri=${encodeURIComponent(process.env.REDIRECT_URI!)}` +
    `&response_type=code` +
    `&state=random_state_string`

  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-24">
      <h1 className="text-4xl font-bold mb-8">じょぎアプリ</h1>
      <Link 
        href={authUrl}
        className="bg-blue-600 text-white px-6 py-3 rounded-lg hover:bg-blue-700 transition"
      >
        Discordでログインして利用開始
      </Link>
    </main>
  )
}
```

## Step 4: コールバックハンドラの実装

認証サーバーから戻ってきた時に実行されるAPIルートを作成します。
ここで認可コードとアクセストークンを交換し、Cookieに保存します。

`app/api/auth/callback/route.ts`:

```ts
import { NextResponse } from 'next/server'
import { cookies } from 'next/headers'

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url)
  const code = searchParams.get('code')

  if (!code) {
    return NextResponse.json({ error: 'No code provided' }, { status: 400 })
  }

  // トークン交換リクエスト
  const tokenRes = await fetch(`${process.env.NEXT_PUBLIC_AUTH_SERVER_URL}/oauth/token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      grant_type: 'authorization_code',
      code: code,
      redirect_uri: process.env.REDIRECT_URI!,
      client_id: process.env.CLIENT_ID!,
      client_secret: process.env.CLIENT_SECRET!,
    }),
  })

  if (!tokenRes.ok) {
    return NextResponse.json({ error: 'Failed to fetch token' }, { status: 500 })
  }

  const data = await tokenRes.json()
  
  // アクセストークンをCookieに保存 (httpOnly)
  const cookieStore = cookies()
  cookieStore.set('access_token', data.access_token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    maxAge: data.expires_in,
    path: '/',
  })

  // マイページへリダイレクト
  return NextResponse.redirect(new URL('/me', request.url))
}
```

## Step 5: マイページの作成 (保護されたページ)

ログイン後に表示されるページを作成します。
サーバーコンポーネントでCookieからトークンを取得し、認証サーバーのユーザー情報APIを叩きます。

`app/me/page.tsx`:

```tsx
import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'

async function getUser(token: string) {
  const res = await fetch(`${process.env.NEXT_PUBLIC_AUTH_SERVER_URL}/api/user`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  })
  
  if (!res.ok) return null
  return res.json()
}

export default async function MePage() {
  const cookieStore = cookies()
  const token = cookieStore.get('access_token')

  if (!token) {
    redirect('/')
  }

  const user = await getUser(token.value)

  if (!user) {
    return <div>セッションが無効です。再度ログインしてください。</div>
  }

  return (
    <main className="p-24">
      <h1 className="text-2xl font-bold mb-4">ようこそ、{user.username}さん</h1>
      <div className="bg-gray-100 p-6 rounded-lg">
        <p>Discord ID: {user.discord_id}</p>
        <p>会員ステータス: <span className="text-green-600 font-bold">有効</span></p>
      </div>
    </main>
  )
}
```

## テスト実行

```bash
npm run dev
```

1. `http://localhost:3000` にアクセス
2. 「Discordでログイン」ボタンをクリック
3. 認証サーバー (Discord) で承認
4. ローカルホストに戻り、マイページで自分のユーザー名が表示されれば成功です！
