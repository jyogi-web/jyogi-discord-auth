'use client';

import { redirectToLogin } from '@/lib/auth';

export default function LoginButton() {
  return (
    <button
      onClick={redirectToLogin}
      className="bg-indigo-600 hover:bg-indigo-700 text-white font-bold py-3 px-6 rounded-lg transition"
    >
      Discordでログイン
    </button>
  );
}
