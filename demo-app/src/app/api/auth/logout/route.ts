import { NextResponse } from 'next/server';
import { cookies, headers } from 'next/headers';

export async function POST() {
  const headersList = headers();
  const origin = headersList.get('origin');
  const appUrl = process.env.NEXT_PUBLIC_APP_URL || 'http://localhost:3000';

  // Basic CSRF check: verify Origin matches APP_URL
  // Note: For stricter security, use a CSRF token
  if (origin && origin !== appUrl) {
    return NextResponse.json({ error: 'Invalid origin' }, { status: 403 });
  }

  try {
    const cookieStore = await cookies();
    cookieStore.delete('access_token');
    cookieStore.delete('refresh_token');

    return NextResponse.json({ success: true });
  } catch (error) {
    console.error('Failed to logout:', error);
    return NextResponse.json({ error: 'Failed to logout' }, { status: 500 });
  }
}
