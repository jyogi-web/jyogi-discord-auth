# Deployment

This guide explains how to deploy the Jyogi Member Authentication System.

The recommended configuration is as follows:

- **Auth Server**: Google Cloud Run
- **Profile Sync**: Google Cloud Functions

## Google Cloud Run (Auth Server)

### Overview

Cloud Run is a serverless platform for running containerized applications.
Since the Auth Server handles HTTP requests, it is suitable for Cloud Run.

### Setup

#### 1. Entry Point

Configure the Dockerfile to execute `cmd/server/main.go` (Use the `Dockerfile` at the root of the repository).

#### 2. Deployment Script

You can deploy using `scripts/deploy-cloud-run.sh`.

```bash
export GCP_PROJECT_ID="your-project-id"
./scripts/deploy-cloud-run.sh
```

### Environment Variables

Please monitor the following environment variables for the Cloud Run service:

- `DISCORD_CLIENT_ID`
- `DISCORD_CLIENT_SECRET`
- `DISCORD_REDIRECT_URI`
- `DISCORD_GUILD_ID`
- `JWT_SECRET`
- `CORS_ALLOWED_ORIGINS`

For detailed instructions, refer to `docs/deployment-cloud-run.md` in the repository.

---

## Google Cloud Functions (Profile Sync)

### Overview

Since profile synchronization is a batch process executed periodically, the combination of Cloud Functions and Cloud Scheduler is optimal.

### Deployment Steps

Use the scripts in the `deployments/cloud-functions` directory.

```bash
cd deployments/cloud-functions
cp .env.yaml.example .env.yaml
./deploy.sh
```

### Scheduling

Run `setup-scheduler.sh` to set up periodic execution (cron).

```bash
./setup-scheduler.sh
```

For more details, refer to `docs/deployment-functions.md` in the repository.

## Other Deployment Options

- **Docker**: Can be run on any server using `docker-compose.yml`.
- **AWS Lambda**: Configuration examples are available in `deployments/aws-lambda`.
