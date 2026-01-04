# Quick Start (Client Integration)

This is a complete hands-on guide to creating a Next.js project from scratch and integrating "Jyogi Member Auth".
We will use the App Router to implement a secure authentication flow.

## Overview

What we will build in this guide:

1. A Next.js application
2. Login Button (Redirect to Auth Server)
3. Callback Handler (Exchange authorization code for token)
4. Profile Page (Access API using the obtained token)

## Step 1: Create a Next.js Project

First, create a new Next.js project.

```bash
npx create-next-app@latest jyogi-client-demo
# Press Enter for all default settings
cd jyogi-client-demo
```

## Step 2: Configure Environment Variables

Create a `.env.local` file in the project root and set your credentials.
Get these values from the Auth Server administrator.

```bash
# .env.local
NEXT_PUBLIC_AUTH_SERVER_URL="https://your-auth-server-url.com"
CLIENT_ID="your_client_id"
CLIENT_SECRET="your_client_secret"
REDIRECT_URI="http://localhost:3000/api/auth/callback"
```

## Step 3: Create Login Button

Add a login link to the main page.
While generating a `state` parameter is best practice for CSRF protection, we use a static string here for simplicity.

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
      <h1 className="text-4xl font-bold mb-8">Jyogi App</h1>
      <Link 
        href={authUrl}
        className="bg-blue-600 text-white px-6 py-3 rounded-lg hover:bg-blue-700 transition"
      >
        Login with Discord
      </Link>
    </main>
  )
}
```

## Step 4: Implement Callback Handler

Create an API Route to handle the callback from the Auth Server.
Here we exchange the authorization code for an access token and save it in a Cookie.

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

  // Token exchange request
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
  
  // Save access token to Cookie (httpOnly)
  const cookieStore = cookies()
  cookieStore.set('access_token', data.access_token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    maxAge: data.expires_in,
    path: '/',
  })

  // Redirect to profile page
  return NextResponse.redirect(new URL('/me', request.url))
}
```

## Step 5: Create Profile Page (Protected)

Create a page displayed after login.
Retrieve the token from the Cookie in a Server Component and call the Auth Server's User Info API.

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
    return <div>Session is invalid. Please login again.</div>
  }

  return (
    <main className="p-24">
      <h1 className="text-2xl font-bold mb-4">Welcome, {user.username}</h1>
      <div className="bg-gray-100 p-6 rounded-lg">
        <p>Discord ID: {user.discord_id}</p>
        <p>Membership Status: <span className="text-green-600 font-bold">Active</span></p>
      </div>
    </main>
  )
}
```

## Run the Test

```bash
npm run dev
```

1. Access `http://localhost:3000`
2. Click "Login with Discord"
3. Approve in Auth Server (Discord)
4. You will be redirected back to localhost, and if you see your username, it's a success!
