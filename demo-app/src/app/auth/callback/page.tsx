'use client';

import { useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';

export default function CallbackPage() {
  const router = useRouter();
  const searchParams = useSearchParams();

  useEffect(() => {
    const error = searchParams.get('error');

    if (error) {
      // エラー時はホームへリダイレクト
      router.push('/?error=' + error);
    } else {
      // 成功時はダッシュボードへリダイレクト
      router.push('/dashboard');
    }
  }, [router, searchParams]);

  return (
    <div className="flex min-h-screen items-center justify-center">
      <p>認証処理中...</p>
    </div>
  );
}
