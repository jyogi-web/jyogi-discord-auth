import Image from 'next/image';
import type { User } from '@/types';

interface MemberListProps {
  members: User[];
}

export default function MemberList({ members }: MemberListProps) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {members.map((member, index) => (
        <div key={member.id || `member-${index}`} className="bg-white shadow rounded-lg p-4">
          <div className="flex items-center gap-3 mb-3">
            {member.avatar_url ? (
              <Image
                src={member.avatar_url}
                alt={`${member.username}のアバター`}
                width={48}
                height={48}
                className="rounded-full"
                priority={false}
              />
            ) : (
              <div className="w-12 h-12 rounded-full bg-gray-300 flex items-center justify-center text-gray-600 font-bold">
                {member.username.charAt(0).toUpperCase()}
              </div>
            )}
            <div>
              <h4 className="font-semibold">{member.username}</h4>
              <p className="text-sm text-gray-600">ID: {member.discord_id}</p>
            </div>
          </div>
          {member.profile && (
            <div className="border-t pt-3 space-y-1">
              {member.profile.real_name && (
                <p className="text-sm"><span className="font-medium">本名:</span> {member.profile.real_name}</p>
              )}
              {member.profile.student_id && (
                <p className="text-sm"><span className="font-medium">学籍番号:</span> {member.profile.student_id}</p>
              )}
              {member.profile.hobbies && (
                <p className="text-sm"><span className="font-medium">趣味:</span> {member.profile.hobbies}</p>
              )}
              {member.profile.what_to_do && (
                <p className="text-sm"><span className="font-medium">やりたいこと:</span> {member.profile.what_to_do}</p>
              )}
              {member.profile.comment && (
                <p className="text-sm"><span className="font-medium">コメント:</span> {member.profile.comment}</p>
              )}
            </div>
          )}
        </div>
      ))}
    </div>
  );
}
