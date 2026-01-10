import { NextResponse } from 'next/server';
import { cookies } from 'next/headers';

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const code = searchParams.get('code');
  const error = searchParams.get('error');

  if (error) {
    return NextResponse.redirect(new URL(`/?error=${error}`, request.url));
  }

  if (!code) {
    return NextResponse.json({ error: 'No code provided' }, { status: 400 });
  }

  const authServerUrl = process.env.NEXT_PUBLIC_AUTH_SERVER_URL;
  const clientId = process.env.CLIENT_ID; // Server-side env var
  const clientSecret = process.env.CLIENT_SECRET; // Server-side env var
  const redirectUri = process.env.REDIRECT_URI; // Server-side env var

  if (!authServerUrl || !clientId || !clientSecret || !redirectUri) {
    console.error('Missing environment variables');
    return NextResponse.json({ error: 'Server configuration error' }, { status: 500 });
  }

  try {
    // トークン交換リクエスト
    const tokenRes = await fetch(`${authServerUrl}/oauth/token`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        grant_type: 'authorization_code',
        code: code,
        redirect_uri: redirectUri,
        client_id: clientId,
        client_secret: clientSecret,
      }),
    });

    if (!tokenRes.ok) {
      const errorData = await tokenRes.json();
      console.error('Token exchange failed:', errorData);
      return NextResponse.redirect(new URL(`/?error=token_exchange_failed`, request.url));
    }

    const data = await tokenRes.json();
    
    // アクセストークンをCookieに保存 (httpOnly)
    // Note: next/headers cookies() is read-only in some contexts, but works in Route Handlers
    const cookieStore = cookies();
    cookieStore.set('access_token', data.access_token, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      maxAge: data.expires_in,
      path: '/',
    });

    // リフレッシュトークンも保存（あれば）
    if (data.refresh_token) {
      cookieStore.set('refresh_token', data.refresh_token, {
        httpOnly: true,
        secure: process.env.NODE_ENV === 'production',
        maxAge: 7 * 24 * 60 * 60, // 7 days (adjust as needed)
        path: '/',
      });
    }

    // ダッシュボードへリダイレクト
    return NextResponse.redirect(new URL('/dashboard', request.url));

  } catch (err) {
    console.error('Callback error:', err);
    return NextResponse.redirect(new URL(`/?error=server_error`, request.url));
  }
}
