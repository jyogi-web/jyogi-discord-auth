#!/usr/bin/env python3
"""
じょぎメンバー認証システム 統合テスト

全フローを包括的にテストします：
1. ヘルスチェック
2. JWT発行・検証フロー
3. JWT更新フロー
4. OAuth2/SSOフロー
5. ログアウトフロー
"""

import json
import sqlite3
import time
import urllib.request
import urllib.parse
import urllib.error
from datetime import datetime, timedelta
import hashlib
import uuid

BASE_URL = "http://localhost:8080"
DB_PATH = "./jyogi_auth.db"

class Colors:
    GREEN = '\033[92m'
    RED = '\033[91m'
    BLUE = '\033[94m'
    YELLOW = '\033[93m'
    END = '\033[0m'

def print_test(name):
    print(f"\n{Colors.BLUE}{'='*60}{Colors.END}")
    print(f"{Colors.BLUE}Test: {name}{Colors.END}")
    print(f"{Colors.BLUE}{'='*60}{Colors.END}")

def print_success(message):
    print(f"{Colors.GREEN}✅ {message}{Colors.END}")

def print_error(message):
    print(f"{Colors.RED}❌ {message}{Colors.END}")

def print_info(message):
    print(f"{Colors.YELLOW}ℹ️  {message}{Colors.END}")

class NoRedirectHandler(urllib.request.HTTPRedirectHandler):
    """リダイレクトを自動フォローしないハンドラー"""
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None

def make_request(url, method="GET", data=None, headers=None, follow_redirects=True):
    """HTTPリクエストを送信"""
    if headers is None:
        headers = {}

    if data is not None:
        if isinstance(data, dict):
            data = urllib.parse.urlencode(data).encode('utf-8')
            headers['Content-Type'] = 'application/x-www-form-urlencoded'
        elif isinstance(data, str):
            data = data.encode('utf-8')

    req = urllib.request.Request(url, data=data, headers=headers, method=method)

    # リダイレクトを自動フォローしない場合
    if not follow_redirects:
        opener = urllib.request.build_opener(NoRedirectHandler)
        try:
            response = opener.open(req)
            body = response.read().decode('utf-8')
            return response.status, body, dict(response.headers)
        except urllib.error.HTTPError as e:
            body = e.read().decode('utf-8')
            return e.code, body, dict(e.headers)
    else:
        try:
            with urllib.request.urlopen(req) as response:
                body = response.read().decode('utf-8')
                return response.status, body, dict(response.headers)
        except urllib.error.HTTPError as e:
            body = e.read().decode('utf-8')
            return e.code, body, dict(e.headers)

def setup_test_data():
    """テストデータをデータベースに準備"""
    print_info("Setting up test data...")

    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()

    # テストユーザーを作成
    user_id = str(uuid.uuid4())
    discord_id = "test_discord_" + str(int(time.time()))
    now = datetime.now().isoformat()

    cursor.execute("""
        INSERT OR REPLACE INTO users (id, discord_id, username, avatar_url, created_at, updated_at, last_login_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    """, (user_id, discord_id, "Test User", "https://example.com/avatar.png", now, now, now))

    # テストセッションを作成
    session_id = str(uuid.uuid4())
    session_token = "test_session_" + str(int(time.time()))
    expires_at = (datetime.now() + timedelta(days=7)).isoformat()

    cursor.execute("""
        INSERT OR REPLACE INTO sessions (id, user_id, token, expires_at, created_at)
        VALUES (?, ?, ?, ?, ?)
    """, (session_id, user_id, session_token, expires_at, now))

    # テストクライアントを作成（OAuth2用）
    client_id = str(uuid.uuid4())
    client_client_id = "test_client_" + str(int(time.time()))
    # bcryptハッシュ（"test_secret"のハッシュ）
    import subprocess
    result = subprocess.run(
        ['python3', '-c', 'import bcrypt; print(bcrypt.hashpw(b"test_secret", bcrypt.gensalt(10)).decode())'],
        capture_output=True,
        text=True
    )
    client_secret = result.stdout.strip() if result.returncode == 0 else "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

    redirect_uris = json.dumps(["http://localhost:3000/callback"])

    cursor.execute("""
        INSERT OR REPLACE INTO client_apps (id, client_id, name, client_secret, redirect_uris, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    """, (client_id, client_client_id, "Test Client", client_secret, redirect_uris, now, now))

    conn.commit()
    conn.close()

    print_success(f"Test data created:")
    print(f"  User ID: {user_id}")
    print(f"  Session Token: {session_token}")
    print(f"  Client ID: {client_client_id}")

    return {
        'user_id': user_id,
        'session_token': session_token,
        'client_id': client_client_id,
        'client_secret': 'test_secret',
        'redirect_uri': 'http://localhost:3000/callback'
    }

def test_health_check():
    """ヘルスチェックをテスト"""
    print_test("Health Check")

    status, body, _ = make_request(f"{BASE_URL}/health")

    if status == 200 and body == "OK":
        print_success("Health check passed")
        return True
    else:
        print_error(f"Health check failed: {status} - {body}")
        return False

def test_jwt_issuance(session_token):
    """JWT発行をテスト"""
    print_test("JWT Issuance")

    # セッションCookieを設定してJWTを発行
    headers = {
        'Cookie': f'session_token={session_token}'
    }

    status, body, _ = make_request(f"{BASE_URL}/token", method="POST", headers=headers)

    if status == 200:
        data = json.loads(body)
        access_token = data.get('access_token')
        token_type = data.get('token_type')
        expires_in = data.get('expires_in')

        print_success(f"JWT issued successfully")
        print(f"  Token Type: {token_type}")
        print(f"  Expires In: {expires_in} seconds")
        print(f"  Access Token: {access_token[:50]}...")

        return access_token
    else:
        print_error(f"JWT issuance failed: {status} - {body}")
        return None

def test_jwt_verification(access_token):
    """JWT検証をテスト"""
    print_test("JWT Verification")

    headers = {
        'Authorization': f'Bearer {access_token}'
    }

    status, body, _ = make_request(f"{BASE_URL}/api/verify", headers=headers)

    if status == 200:
        data = json.loads(body)
        print_success("JWT verification passed")
        print(f"  Valid: {data.get('valid')}")
        print(f"  User ID: {data.get('user_id')}")
        print(f"  Username: {data.get('username')}")
        return True
    else:
        print_error(f"JWT verification failed: {status} - {body}")
        return False

def test_jwt_user_info(access_token):
    """JWT認証でユーザー情報取得をテスト"""
    print_test("JWT User Info Retrieval")

    headers = {
        'Authorization': f'Bearer {access_token}'
    }

    status, body, _ = make_request(f"{BASE_URL}/api/user", headers=headers)

    if status == 200:
        data = json.loads(body)
        print_success("User info retrieved successfully")
        print(f"  ID: {data.get('id')}")
        print(f"  Discord ID: {data.get('discord_id')}")
        print(f"  Username: {data.get('username')}")
        return True
    else:
        print_error(f"User info retrieval failed: {status} - {body}")
        return False

def test_jwt_refresh(access_token):
    """JWT更新をテスト"""
    print_test("JWT Refresh")

    headers = {
        'Authorization': f'Bearer {access_token}'
    }

    status, body, _ = make_request(f"{BASE_URL}/token/refresh", method="POST", headers=headers)

    if status == 200:
        data = json.loads(body)
        new_access_token = data.get('access_token')
        print_success("JWT refreshed successfully")
        print(f"  New Access Token: {new_access_token[:50]}...")
        return new_access_token
    else:
        print_error(f"JWT refresh failed: {status} - {body}")
        return None

def test_oauth2_flow(test_data):
    """OAuth2/SSOフローをテスト"""
    print_test("OAuth2/SSO Flow")

    # Step 1: 認可リクエスト（セッションCookie付き）
    params = {
        'client_id': test_data['client_id'],
        'redirect_uri': test_data['redirect_uri'],
        'response_type': 'code',
        'state': 'test_state_123'
    }

    headers = {
        'Cookie': f'session_token={test_data["session_token"]}'
    }

    url = f"{BASE_URL}/oauth/authorize?" + urllib.parse.urlencode(params)
    status, body, response_headers = make_request(url, headers=headers, follow_redirects=False)

    if status == 302:
        location = response_headers.get('Location', '')
        parsed = urllib.parse.urlparse(location)
        query_params = urllib.parse.parse_qs(parsed.query)

        auth_code = query_params.get('code', [None])[0]
        state = query_params.get('state', [None])[0]

        if auth_code and state == 'test_state_123':
            print_success(f"Authorization successful")
            print(f"  Auth Code: {auth_code[:30]}...")
            print(f"  State: {state}")

            # Step 2: トークン交換
            token_data = {
                'grant_type': 'authorization_code',
                'code': auth_code,
                'client_id': test_data['client_id'],
                'client_secret': test_data['client_secret'],
                'redirect_uri': test_data['redirect_uri']
            }

            status, body, _ = make_request(f"{BASE_URL}/oauth/token", method="POST", data=token_data)

            if status == 200:
                data = json.loads(body)
                print_success("Token exchange successful")
                print(f"  Access Token: {data.get('access_token')[:50]}...")
                print(f"  Token Type: {data.get('token_type')}")
                print(f"  Expires In: {data.get('expires_in')} seconds")
                return True
            else:
                print_error(f"Token exchange failed: {status} - {body}")
                return False
        else:
            print_error(f"Authorization code or state missing")
            return False
    else:
        print_error(f"Authorization failed: {status} - {body}")
        return False

def test_logout(session_token):
    """ログアウトをテスト"""
    print_test("Logout Flow")

    headers = {
        'Cookie': f'session_token={session_token}'
    }

    status, body, _ = make_request(f"{BASE_URL}/auth/logout", method="POST", headers=headers)

    if status == 200:
        print_success("Logout successful")

        # ログアウト後にJWT発行を試みる（失敗するはず）
        status2, body2, _ = make_request(f"{BASE_URL}/token", method="POST", headers=headers)

        if status2 == 401:
            print_success("Post-logout JWT issuance correctly denied")
            return True
        else:
            print_error(f"Post-logout JWT issuance should have failed: {status2}")
            return False
    else:
        print_error(f"Logout failed: {status} - {body}")
        return False

def main():
    """メインテスト実行"""
    print(f"\n{Colors.BLUE}{'='*60}")
    print("じょぎメンバー認証システム 統合テスト")
    print(f"{'='*60}{Colors.END}\n")

    # テストデータ準備
    test_data = setup_test_data()

    results = []

    # テスト実行
    results.append(("Health Check", test_health_check()))

    access_token = test_jwt_issuance(test_data['session_token'])
    results.append(("JWT Issuance", access_token is not None))

    if access_token:
        results.append(("JWT Verification", test_jwt_verification(access_token)))
        results.append(("JWT User Info", test_jwt_user_info(access_token)))

        new_token = test_jwt_refresh(access_token)
        results.append(("JWT Refresh", new_token is not None))

    results.append(("OAuth2/SSO Flow", test_oauth2_flow(test_data)))
    results.append(("Logout Flow", test_logout(test_data['session_token'])))

    # 結果サマリー
    print(f"\n{Colors.BLUE}{'='*60}")
    print("Test Summary")
    print(f"{'='*60}{Colors.END}\n")

    passed = sum(1 for _, result in results if result)
    total = len(results)

    for name, result in results:
        status = f"{Colors.GREEN}PASS{Colors.END}" if result else f"{Colors.RED}FAIL{Colors.END}"
        print(f"  {name}: {status}")

    print(f"\n{Colors.BLUE}Total: {passed}/{total} tests passed{Colors.END}")

    if passed == total:
        print(f"\n{Colors.GREEN}✅ All tests passed!{Colors.END}\n")
        return 0
    else:
        print(f"\n{Colors.RED}❌ Some tests failed{Colors.END}\n")
        return 1

if __name__ == "__main__":
    exit(main())
