import { NextResponse } from 'next/server';
import { cookies, headers } from 'next/headers';

export async function POST() {
  const headersList = await headers();
  const origin = headersList.get('origin');
  const appUrl = process.env.NEXT_PUBLIC_APP_URL || 'http://localhost:3000';
  
  // Basic CSRF check: verify Origin matches APP_URL
  // Note: For stricter security, use a CSRF token
  // Normalize appUrl to ensure we are comparing origins
  let allowedOrigin = '';
  try {
    allowedOrigin = new URL(appUrl).origin;
  } catch {
    console.error('Invalid NEXT_PUBLIC_APP_URL configuration');
    return NextResponse.json({ error: 'Server configuration error' }, { status: 500 });
  }

  if (origin && origin !== allowedOrigin) {
    return NextResponse.json({ error: 'Invalid origin' }, { status: 403 });
  }

  try {
    const cookieStore = await cookies();
    // Delete cookies with the same options as they were set
    // Note: path is crucial for successful deletion
    cookieStore.delete({ name: 'access_token', path: '/' });
    cookieStore.delete({ name: 'refresh_token', path: '/' });

    return NextResponse.json({ success: true });
  } catch (error) {
    console.error('Failed to logout:', error);
    return NextResponse.json({ error: 'Failed to logout' }, { status: 500 });
  }
}
