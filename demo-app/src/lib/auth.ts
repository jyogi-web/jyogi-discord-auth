const AUTH_SERVER_URL = process.env.NEXT_PUBLIC_AUTH_SERVER_URL || 'http://localhost:8080';
const APP_URL = process.env.NEXT_PUBLIC_APP_URL || 'http://localhost:3000';

export function redirectToLogin() {
  // redirect_uriを指定して認証サーバーにリダイレクト
  const redirectUri = `${APP_URL}/auth/callback`;
  window.location.href = `${AUTH_SERVER_URL}/auth/login?redirect_uri=${encodeURIComponent(redirectUri)}`;
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
