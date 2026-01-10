import Image from 'next/image';
import type { User } from '@/types';

interface UserProfileProps {
  user: User;
}

export default function UserProfile({ user }: UserProfileProps) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
      {/* 基本情報 */}
      <div className="bg-white shadow rounded-lg p-6">
        <div className="flex items-center gap-4 mb-4">
          {user.avatar_url ? (
            <Image
              src={user.avatar_url}
              alt={`${user.username}のアバター`}
              width={80}
              height={80}
              className="rounded-full"
              priority={true}
            />
          ) : (
            <div className="w-20 h-20 rounded-full bg-gray-300 flex items-center justify-center text-gray-600 font-bold text-2xl">
              {user.username.charAt(0).toUpperCase()}
            </div>
          )}
          <div>
            <h3 className="text-xl font-bold">{user.display_name || user.username}</h3>
            <p className="text-gray-600">@{user.username}</p>
          </div>
        </div>

        <dl className="space-y-2">
          <div>
            <dt className="text-sm text-gray-500">Discord ID</dt>
            <dd>{user.discord_id}</dd>
          </div>
          {user.last_login_at && (
            <div>
              <dt className="text-sm text-gray-500">最終ログイン</dt>
              <dd>{new Date(user.last_login_at).toLocaleString('ja-JP')}</dd>
            </div>
          )}
          {user.guild_nickname && (
             <div>
               <dt className="text-sm text-gray-500">サーバーニックネーム</dt>
               <dd>{user.guild_nickname}</dd>
             </div>
          )}
          {user.joined_at && (
            <div>
              <dt className="text-sm text-gray-500">サーバー参加日</dt>
              <dd>{new Date(user.joined_at).toLocaleString('ja-JP')}</dd>
            </div>
          )}
          {user.guild_roles && user.guild_roles.length > 0 && (
            <div>
              <dt className="text-sm text-gray-500">ロール</dt>
              <dd className="flex flex-wrap gap-2">
                {user.guild_roles.map((roleId) => (
                  <span
                    key={roleId}
                    className="inline-block px-2 py-1 text-xs rounded bg-blue-100 text-blue-800"
                  >
                    {roleId}
                  </span>
                ))}
              </dd>
            </div>
          )}
        </dl>
      </div>

      {/* じょぎプロフィール情報 */}
      {user.profile && (
        <div className="bg-white shadow rounded-lg p-6">
          <h3 className="text-lg font-bold mb-4 border-b pb-2">プロフィール情報</h3>
          <dl className="space-y-3">
            <div>
              <dt className="text-sm text-gray-500">氏名</dt>
              <dd className="font-medium">{user.profile.real_name || '-'}</dd>
            </div>
            <div>
              <dt className="text-sm text-gray-500">学籍番号</dt>
              <dd>{user.profile.student_id || '-'}</dd>
            </div>
            <div>
              <dt className="text-sm text-gray-500">趣味</dt>
              <dd>{user.profile.hobbies || '-'}</dd>
            </div>
            <div>
              <dt className="text-sm text-gray-500">やりたいこと</dt>
              <dd>{user.profile.what_to_do || '-'}</dd>
            </div>
            <div>
              <dt className="text-sm text-gray-500">ひとこと</dt>
              <dd className="text-gray-700 whitespace-pre-wrap">{user.profile.comment || '-'}</dd>
            </div>
          </dl>
        </div>
      )}
    </div>
  );
}