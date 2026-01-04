# Architecture

This document describes the architecture and project structure of the Jyogi Member Authentication System.

## Overview

1. **Identity Provider (IdP)**: Discord (User information, server membership management)
2. **Auth Server (Jyogi Auth)**: Performs Discord OAuth2, determines if the user is a "Jyogi member", and issues a unique access token (JWT).
3. **Client Apps (Internal Tools)**: Applications that delegate login to "Jyogi Auth".

## Technology Stack

- **Language**: Go 1.23+
- **Database**: SQLite (Designed to be migratable to PostgreSQL in the future)
- **Authentication**: Discord OAuth2, JWT
- **Deployment**: Google Cloud Run (Auth Server), Google Cloud Functions (Profile Sync)

## Project Structure

```
jyogi-discord-auth/
├── cmd/
│   ├── server/          # Main server entry point
│   ├── sync-profiles/   # Profile sync tool (CLI)
│   └── sync-profiles-fn/ # Profile sync Function (HTTP)
├── deployments/
│   ├── cloud-functions/ # Google Cloud Functions configuration
│   └── aws-lambda/      # AWS Lambda configuration
├── internal/
│   ├── domain/          # Domain models (User, Profile, Session, etc.)
│   ├── repository/      # Data access layer
│   ├── service/         # Business logic
│   ├── handler/         # HTTP handlers
│   ├── middleware/      # HTTP middleware
│   └── config/          # Configuration management
├── pkg/
│   ├── discord/         # Discord API client, profile parser
│   ├── auth/            # Client authentication
│   └── jwt/             # JWT utilities
├── web/
│   ├── templates/       # HTML templates
│   └── static/          # Static files
├── migrations/          # Database migrations
├── tests/               # Tests
└── scripts/             # Development and operation scripts
```
