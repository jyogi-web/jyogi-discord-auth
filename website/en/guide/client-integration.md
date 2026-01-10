# Quick Start (Client Integration)

This is a complete hands-on guide to building a Next.js project from scratch and integrating "Jyogi Member Auth".
We will use the App Router to implement a secure authentication flow.

## Overview

> [!TIP]
> A complete working example corresponding to this guide can be found in the `demo-app/` directory of this repository.
> Feel free to refer to it while following this guide.

In this guide, we will build:

1. A Next.js application
2. A Login button (redirects to the Auth Server)
3. A Callback handler (exchanges auth code for a token)
4. A Profile page (accesses API using the obtained token)

### Authentication Flow

```mermaid
sequenceDiagram
    participant User as User
    participant Client as Client App
    participant Auth as Auth Server (Jyogi Auth)
    participant Discord as Discord

    User->>Client: Clicks Login Button
    Client->>Auth: Redirect (GET /oauth/authorize)
    Auth->>User: Show Login Screen (if not logged in)
    User->>Auth: Login with Discord
    Auth->>Discord: OAuth2 Integration
    Discord-->>Auth: Returns User Info
    Auth->>Auth: Checks Jyogi Membership
    Auth-->>Client: Redirect (callback?code=...)
    Client->>Auth: Exchange Token (POST /oauth/token)
    Auth-->>Client: Returns Access Token
    Client->>Auth: Get User Info (GET /api/user)
    Auth-->>Client: Returns User Info & Profile
    Client->>User: Displays My Page
```

## Step 1: Create Next.js Project

First, create a new Next.js project.

```bash
npx create-next-app@latest jyogi-client-demo
# Default settings (Enter) are fine
cd jyogi-client-demo
```

## Step 2: Configure Environment Variables

Create a `.env.local` file in the project root and set the authentication credentials.
Get these values from the Auth Server administrator.

```bash
# .env.local
NEXT_PUBLIC_AUTH_SERVER_URL="https://your-auth-server-url.com"
CLIENT_ID="your_client_id"
CLIENT_SECRET="your_client_secret"
REDIRECT_URI="http://localhost:3000/api/auth/callback"
```

## Step 3: Create Login Button

Add a login link to the top page.
Using a generated `state` parameter is best practice for CSRF protection, but we'll use a fixed value here for simplicity.

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

Create an API route that runs when returning from the Auth Server.
Here, we exchange the auth code for an access token and save it in a Cookie.

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
  const cookieStore = await cookies()
  cookieStore.set('access_token', data.access_token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    maxAge: data.expires_in,
    path: '/',
  })

  // Redirect to My Page
  return NextResponse.redirect(new URL('/me', request.url))
}
```

## Step 5: Create My Page (Protected Page)

Create a page displayed after login.
Fetch the token from the Cookie in a Server Component and hit the Auth Server's user info API.

`app/me/page.tsx`:

```tsx
import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'

// API Response Type Definition
type UserProfile = {
  id: string
  discord_id: string
  username: string
  display_name: string
  avatar_url: string
  guild_roles?: string[]
  profile?: {
    real_name?: string
    student_id?: string
    hobbies?: string
    what_to_do?: string
    comment?: string
  }
}

async function getUser(token: string): Promise<UserProfile | null> {
  const res = await fetch(`${process.env.NEXT_PUBLIC_AUTH_SERVER_URL}/api/user`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    cache: 'no-store' // Always fetch fresh data
  })
  
  if (!res.ok) return null
  return res.json()
}

export default async function MePage() {
  const cookieStore = await cookies()
  const token = cookieStore.get('access_token')

  if (!token) {
    redirect('/')
  }

  const user = await getUser(token.value)

  if (!user) {
    return <div>Session invalid. Please login again.</div>
  }

  return (
    <main className="p-24">
      <div className="flex items-center gap-4 mb-8">
        {user.avatar_url && (
          <img src={user.avatar_url} alt={user.username} className="w-16 h-16 rounded-full" />
        )}
        <div>
          <h1 className="text-2xl font-bold">{user.display_name}</h1>
          <p className="text-gray-500">@{user.username}</p>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="bg-white p-6 rounded-lg shadow">
          <h2 className="text-xl font-bold mb-4">Account Info</h2>
          <dl className="space-y-2">
            <div>
              <dt className="text-sm text-gray-500">Discord ID</dt>
              <dd>{user.discord_id}</dd>
            </div>
            <div>
              <dt className="text-sm text-gray-500">Membership Status</dt>
              <dd className="text-green-600 font-bold">Active</dd>
            </div>
          </dl>
        </div>

        {user.profile && (
          <div className="bg-white p-6 rounded-lg shadow">
            <h2 className="text-xl font-bold mb-4">Profile (Synced)</h2>
            <dl className="space-y-2">
              <div>
                <dt className="text-sm text-gray-500">Real Name</dt>
                <dd>{user.profile.real_name || '-'}</dd>
              </div>
              <div>
                <dt className="text-sm text-gray-500">Student ID</dt>
                <dd>{user.profile.student_id || '-'}</dd>
              </div>
              <div>
                <dt className="text-sm text-gray-500">Ambitions</dt>
                <dd>{user.profile.what_to_do || '-'}</dd>
              </div>
            </dl>
          </div>
        )}
      </div>
    </main>
  )
}
```

## Implementation in Other Languages

### Python (Requests)

```python
import requests

def get_user_profile(access_token):
    url = "https://your-auth-server.com/api/user"
    headers = {
        "Authorization": f"Bearer {access_token}"
    }
    
    response = requests.get(url, headers=headers)
    
    if response.status_code == 200:
        return response.json()
    else:
        print(f"Error: {response.status_code}")
        return None

# Example usage
profile = get_user_profile("YOUR_ACCESS_TOKEN")
print(profile['username'])
```

### Node.js (Axios)

```javascript
const axios = require('axios');

async function getUserProfile(accessToken) {
  try {
    const response = await axios.get('https://your-auth-server.com/api/user', {
      headers: {
        Authorization: `Bearer ${accessToken}`
      }
    });
    return response.data;
  } catch (error) {
    console.error('Error fetching profile:', error.response?.status);
    return null;
  }
}
```

### Ruby on Rails (Faraday)

```ruby
# app/controllers/auth_controller.rb
class AuthController < ApplicationController
  def callback
    # 1. Exchange authorization code for access token
    response = Faraday.post("#{ENV['AUTH_SERVER_URL']}/oauth/token") do |req|
      req.headers['Content-Type'] = 'application/x-www-form-urlencoded'
      req.body = URI.encode_www_form({
        grant_type: 'authorization_code',
        code: params[:code],
        client_id: ENV['CLIENT_ID'],
        client_secret: ENV['CLIENT_SECRET'],
        redirect_uri: ENV['REDIRECT_URI']
      })
    end

    unless response.success?
      return redirect_to root_path, alert: 'Auth failed'
    end

    access_token = JSON.parse(response.body)['access_token']

    # 2. Get user info and profile
    user_res = Faraday.get("#{ENV['AUTH_SERVER_URL']}/api/user") do |req|
      req.headers['Authorization'] = "Bearer #{access_token}"
    end

    user_info = JSON.parse(user_res.body)
    session[:user] = user_info
    redirect_to dashboard_path
  end
end
```

## Error Handling

When using the API, it is recommended to handle the following error codes:

| Status Code | Meaning | Solution |
| :--- | :--- | :--- |
| `401 Unauthorized` | Token invalid or expired | Prompt user to re-login or use refresh token. |
| `403 Forbidden` | No access permission | User might not be a Jyogi member. |
| `500 Internal Server Error` | Server error | Contact the administrator. |