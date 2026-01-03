const AUTH_SERVER_URL = process.env.NEXT_PUBLIC_AUTH_SERVER_URL || 'http://localhost:8080';

export const api = {
  // 現在のユーザー情報を取得
  async getCurrentUser() {
    const res = await fetch(`${AUTH_SERVER_URL}/api/me`, {
      credentials: 'include',
    });
    if (!res.ok) throw new Error('Failed to fetch user');
    return res.json();
  },

  // メンバー一覧を取得
  async getMembers() {
    const res = await fetch(`${AUTH_SERVER_URL}/api/members`, {
      credentials: 'include',
    });
    if (!res.ok) throw new Error('Failed to fetch members');
    return res.json();
  },

  // ログアウト
  async logout() {
    const res = await fetch(`${AUTH_SERVER_URL}/auth/logout`, {
      method: 'POST',
      credentials: 'include',
    });
    if (!res.ok) throw new Error('Failed to logout');
    return res.json();
  },
};
