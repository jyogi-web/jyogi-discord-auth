'use client';

import { logout } from '@/lib/auth';
import type { User } from '@/types';

interface HeaderProps {
  user: User | null;
}

export default function Header({ user }: HeaderProps) {
  return (
    <header className="bg-white shadow">
      <nav className="container mx-auto px-4 py-4 flex justify-between items-center">
        <h1 className="text-xl font-bold">じょぎ認証デモ</h1>
        {user && (
          <div className="flex items-center gap-4">
            <span className="text-gray-700">{user.username}</span>
            <button
              onClick={logout}
              className="bg-red-500 hover:bg-red-600 text-white px-4 py-2 rounded transition"
            >
              ログアウト
            </button>
          </div>
        )}
      </nav>
    </header>
  );
}
