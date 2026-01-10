const AUTH_SERVER_URL = process.env.NEXT_PUBLIC_AUTH_SERVER_URL || 'http://localhost:8080';
const APP_URL = process.env.NEXT_PUBLIC_APP_URL || 'http://localhost:3000';

export function redirectToLogin() {
  // CSRF対策: stateを生成して保存
  const state = Math.random().toString(36).substring(7);
  sessionStorage.setItem('oauth_state', state);

  // redirect_uriを指定して認証サーバーにリダイレクト
  const redirectUri = `${APP_URL}/api/auth/callback`; // Note: Using the new API route
  // 本来はAuth Serverの /oauth/authorize エンドポイントを叩くべきだが、
  // このデモではAuth Server側の便利機能 /auth/login を使用しているため、
  // stateパラメータの扱いはAuth Serverの実装に依存する。
  // Auth Serverの /auth/login は内部でstateを生成してしまうため、
  // クライアント側で生成したstateを検証するには、/oauth/authorizeを直接叩く必要がある。
  
  // ここではより標準的な OAuth2 フローとして /oauth/authorize を直接構築する例に変更
  const clientId = process.env.NEXT_PUBLIC_CLIENT_ID; // Client-side env var needed
  if (!clientId) {
      console.error("NEXT_PUBLIC_CLIENT_ID is not set");
      // Fallback to the simplified login endpoint if ID is missing (dev mode)
      const devRedirectUri = `${APP_URL}/auth/callback`;
      window.location.href = `${AUTH_SERVER_URL}/auth/login?redirect_uri=${encodeURIComponent(devRedirectUri)}`;
      return;
  }

  const authUrl = `${AUTH_SERVER_URL}/oauth/authorize?` + 
    `client_id=${clientId}&` +
    `redirect_uri=${encodeURIComponent(redirectUri)}&` +
    `response_type=code&` +
    `state=${state}`;

  window.location.href = authUrl;
}

export async function logout() {
  try {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 5000);

    await fetch('/api/auth/logout', { 
      method: 'POST',
      signal: controller.signal
    });
    clearTimeout(timeoutId);
    
    window.location.href = '/';
  } catch (error) {
    console.error('Logout failed:', error);
    // Force redirect even if logout fails
    window.location.href = '/';
  }
}
