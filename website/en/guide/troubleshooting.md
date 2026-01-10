# Troubleshooting

Common errors and their solutions.

## Authentication Issues

### "access_denied" Error

Occurs when the user clicks "Cancel" on the Discord authorization screen or authorization is denied by Discord.
If this happens unexpectedly, check if `DISCORD_CLIENT_ID` is correct.

### "redirect_uri_mismatch" Error

The redirect URI registered in the Discord Developer Portal does not match the `redirect_uri` sent by the application.
- Check `DISCORD_REDIRECT_URI` in `.env`.
- Ensure the redirect URI in the Developer Portal's "OAuth2" settings matches exactly (including trailing slashes).

### JWT Verification Error (401 Unauthorized)

Ensure the `Authorization: Bearer <token>` header is correctly set when making API requests.
The token might also be expired (default 24 hours). Try refreshing it using the `/token/refresh` endpoint.

## Database & Startup Issues

### Database Connection Error

```
Failed to connect to TiDB ...
```

- **Development (SQLite)**: Check folder write permissions.
- **Production (TiDB)**: Check `TIDB_DB_HOST`, `TIDB_DB_USERNAME`, `TIDB_DB_PASSWORD`. If connecting from Cloud Run, check VPC connector settings and IP restrictions.

### Port Conflict

```
bind: address already in use
```

The port specified by `SERVER_PORT` (default 8080) is already in use.
Stop the other process or change the port number in `.env`.

## CORS Errors

If you see CORS-related errors in the browser console:

- Check if `CORS_ALLOWED_ORIGINS` in `.env` includes the request origin (e.g., `http://localhost:3000`).
- It must match exactly, including protocol (http/https) and port number.
