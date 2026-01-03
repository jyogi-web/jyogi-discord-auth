'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Header from '@/components/Header';
import UserProfile from '@/components/UserProfile';
import MemberList from '@/components/MemberList';
import { api } from '@/lib/api';
import type { User, MembersResponse } from '@/types';

export default function Dashboard() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [members, setMembers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchData() {
      try {
        const [userData, membersData]: [User, MembersResponse] = await Promise.all([
          api.getCurrentUser(),
          api.getMembers(),
        ]);
        setUser(userData);
        setMembers(membersData.members);
      } catch (error) {
        console.error('Failed to fetch data:', error);
        router.push('/');
      } finally {
        setLoading(false);
      }
    }
    fetchData();
  }, [router]);

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <p>読み込み中...</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <Header user={user} />
      <main className="container mx-auto px-4 py-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold mb-4">ダッシュボード</h1>
          {user && <UserProfile user={user} />}
        </div>
        <div>
          <h2 className="text-2xl font-bold mb-4">じょぎメンバー一覧</h2>
          <MemberList members={members} />
        </div>
      </main>
    </div>
  );
}
