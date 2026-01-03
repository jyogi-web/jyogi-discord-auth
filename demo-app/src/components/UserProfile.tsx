import Image from 'next/image';
import type { User } from '@/types';

interface UserProfileProps {
  user: User;
}

export default function UserProfile({ user }: UserProfileProps) {
  return (
    <div className="bg-white shadow rounded-lg p-6 flex items-center gap-4">
      {user.avatar_url ? (
        <Image
          src={user.avatar_url}
          alt={`${user.username}のアバター`}
          width={64}
          height={64}
          className="rounded-full"
          priority={true}
        />
      ) : (
        <div className="w-16 h-16 rounded-full bg-gray-300 flex items-center justify-center text-gray-600 font-bold text-xl">
          {user.username.charAt(0).toUpperCase()}
        </div>
      )}
      <div>
        <h3 className="text-xl font-bold">{user.username}</h3>
        <p className="text-gray-600">Discord ID: {user.discord_id}</p>
        <p className="text-sm text-gray-500">
          最終ログイン: {new Date(user.last_login_at).toLocaleString('ja-JP')}
        </p>
      </div>
    </div>
  );
}
