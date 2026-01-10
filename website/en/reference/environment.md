# Environment Variables Reference

List of environment variables for configuring the system.
Set these in a `.env` file for development, or as service environment variables for production (e.g., Cloud Run).

## Required Configuration

| Variable | Description | Example |
| :--- | :--- | :--- |
| `DISCORD_CLIENT_ID` | Client ID obtained from Discord Developer Portal | `123456789012345678` |
| `DISCORD_CLIENT_SECRET` | Client Secret obtained from Discord Developer Portal | `abcdefghijklmnopqrstuvwxyz` |
| `DISCORD_REDIRECT_URI` | OAuth2 Callback URL | `http://localhost:8080/auth/callback` |
| `DISCORD_GUILD_ID` | Target Discord Server ID | `987654321098765432` |
| `JWT_SECRET` | Secret key for JWT signing (min 32 chars) | `your-secure-random-string-minimum-32-chars` |

## Profile Sync Configuration

Required when using the profile synchronization feature.

| Variable | Description | Example |
| :--- | :--- | :--- |
| `DISCORD_BOT_TOKEN` | Discord Bot Token | `MTA...` |
| `DISCORD_PROFILE_CHANNEL` | Introduction Channel ID | `123456789012345678` |

## Server & DB Configuration

| Variable | Description | Default |
| :--- | :--- | :--- |
| `SERVER_PORT` | Port the server listens on | `8080` |
| `ENV` | Execution environment (`development` / `production`) | `development` |
| `DATABASE_PATH` | Path to SQLite database file (for development) | `./jyogi_auth.db` |
| `HTTPS_ONLY` | Enforce HTTPS (`true` / `false`) | `false` |
| `CORS_ALLOWED_ORIGINS` | Allowed origins for CORS (comma separated) | `http://localhost:3000` |

## Cloud Run / TiDB Configuration (Production)

| Variable | Description |
| :--- | :--- |
| `GCP_PROJECT_ID` | Google Cloud Project ID |
| `GCP_REGION` | Deployment Region (e.g., `asia-northeast1`) |
| `TIDB_DB_HOST` | TiDB Hostname |
| `TIDB_DB_PORT` | TiDB Port (default: `4000`) |
| `TIDB_DB_USERNAME` | TiDB Username |
| `TIDB_DB_PASSWORD` | TiDB Password |
| `TIDB_DB_DATABASE` | Database Name |
| `TIDB_DISABLE_TLS` | Disable TLS connection (`true` / `false`) |

## Security Best Practices

### Secret Management

- **Development**: Use a `.env` file and ensure it is added to `.gitignore`.
- **Production**: Set environment variables directly or use a secret management service like GCP Secret Manager.
- **Generating Secrets**: You can generate secure secrets like `JWT_SECRET` using the following command:
  ```bash
  openssl rand -base64 32
  ```

### Mandatory Production Settings

Ensure the following settings are enabled in production (`ENV=production`):

- `HTTPS_ONLY=true`: Enforce HTTPS connections.
- `CORS_ALLOWED_ORIGINS`: Restrict to production domains only.
