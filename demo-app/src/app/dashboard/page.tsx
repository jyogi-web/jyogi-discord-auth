import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';
import Header from '@/components/Header';
import UserProfile from '@/components/UserProfile';
import MemberList from '@/components/MemberList';
import type { User, MembersResponse } from '@/types';

async function getData(token: string): Promise<{ user: User | null; members: User[] }> {
  // Use server-side environment variable if available, fallback to public one if needed (e.g. local dev consistency)
  const authServerUrl = process.env.AUTH_SERVER_URL || process.env.NEXT_PUBLIC_AUTH_SERVER_URL;
  
  try {
    const [userRes, membersRes] = await Promise.all([
      fetch(`${authServerUrl}/api/user`, {
        headers: { Authorization: `Bearer ${token}` },
        cache: 'no-store',
      }),
      fetch(`${authServerUrl}/api/members`, {
        headers: { Authorization: `Bearer ${token}` }, // Assuming api/members also accepts Bearer
        cache: 'no-store',
      }),
    ]);

    if (!userRes.ok || !membersRes.ok) {
        console.error("API Error", userRes.status, membersRes.status);
        return { user: null, members: [] };
    }

    const user = await userRes.json();
    const membersData: MembersResponse = await membersRes.json();
    
    return { user, members: membersData.members };
  } catch (error) {
    console.error('Fetch error:', error);
    return { user: null, members: [] };
  }
}

export default async function Dashboard() {
  const cookieStore = await cookies();
  const token = cookieStore.get('access_token');

  if (!token) {
    redirect('/');
  }

  const { user, members } = await getData(token.value);

  if (!user) {
      // Token might be expired
      return (
          <div className="flex min-h-screen items-center justify-center flex-col gap-4">
              <p>セッションが無効です。</p>
              <a href="/" className="text-blue-500 hover:underline">トップへ戻る</a>
          </div>
      );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <Header user={user} />
      <main className="container mx-auto px-4 py-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold mb-4">ダッシュボード</h1>
          <UserProfile user={user} />
        </div>
        <div>
          <h2 className="text-2xl font-bold mb-4">じょぎメンバー一覧</h2>
          <MemberList members={members} />
        </div>
      </main>
    </div>
  );
}
