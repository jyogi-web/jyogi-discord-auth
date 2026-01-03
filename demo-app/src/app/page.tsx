import LoginButton from '@/components/LoginButton';

export default function Home() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-24">
      <div className="text-center">
        <h1 className="text-4xl font-bold mb-4">じょぎ認証デモ</h1>
        <p className="text-gray-600 mb-8">
          Discordアカウントでログインして、じょぎメンバー専用機能をお試しください。
        </p>
        <LoginButton />
      </div>
    </main>
  );
}
