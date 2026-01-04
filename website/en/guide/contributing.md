# Getting Started

This guide explains how to set up the development environment for the Jyogi Member Authentication System.

## Prerequisites

1. Go 1.23 or higher installed
2. Application created in Discord Developer Portal
3. Server ID of the Jyogi Discord server obtained

## 1. Clone the Repository

```bash
git clone https://github.com/jyogi-web/jyogi-discord-auth.git
cd jyogi-discord-auth
```

## 2. Install Dependencies

```bash
go mod download
```

## 3. Configure Environment Variables

Create a `.env` file by copying `.env.example` and set the necessary environment variables:

```bash
cp .env.example .env
```

Edit the `.env` file and set the following values:

- `DISCORD_CLIENT_ID`: Client ID obtained from Discord Developer Portal
- `DISCORD_CLIENT_SECRET`: Client Secret obtained from Discord Developer Portal
- `DISCORD_REDIRECT_URI`: OAuth2 Redirect URI (`http://localhost:8080/auth/callback` for local development)
- `DISCORD_GUILD_ID`: Server ID of the Jyogi Discord server
- `DISCORD_BOT_TOKEN`: Discord Bot Token (for profile sync)
- `DISCORD_PROFILE_CHANNEL`: Introduction channel ID
- `JWT_SECRET`: Secret key for JWT signing (at least 32 characters)

## 4. Database Migration

```bash
./scripts/migrate.sh
```

## 5. Start the Server

```bash
go run cmd/server/main.go
```

Once the server starts, access `http://localhost:8080` in your browser.

## Development with Docker

Using Docker simplifies environment setup and provides a unified development environment for the team.

### Prerequisites

- Docker & Docker Compose installed

### Launch

```bash
# Configure environment variables
cp .env.example .env
# Edit .env file and set necessary values

# Build & Launch with Docker Compose
docker-compose up -d

# Check logs
docker-compose logs -f
```

### Makefile

You can execute frequent tasks using the make command.

```bash
make help        # Show help
make run         # Run locally
make docker-up   # Run with Docker
```
