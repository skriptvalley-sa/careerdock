# CareerDock — Deployment Strategy

> **Version:** 1.0
> **Status:** Draft (Phase 6)
> **Last updated:** 2026-03-12
> **Depends on:** [ARCHITECTURE.md](./ARCHITECTURE.md), [CODE-STRUCTURE.md](./CODE-STRUCTURE.md), [SECURITY.md](./SECURITY.md)

---

## 1. Overview

CareerDock runs on AWS with a single-EC2 Docker Compose topology for MVP. This document covers:

1. **Production infrastructure setup** — EC2, RDS, S3, CloudFront, DNS, SSL
2. **CI/CD pipeline** — GitHub Actions from tag push to production deploy
3. **Day-to-day operations** — deployments, migrations, rollbacks, health checks
4. **Scaling roadmap** — when and how to upgrade infrastructure

### 1.1 Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Container registry | AWS ECR | AWS-native, private by default, IAM integration, 500MB free |
| Infrastructure-as-Code | Deferred (manual setup + docs) | Solo founder — speed over reproducibility for MVP |
| Staging environment | Deferred (local Docker Compose for testing) | Saves ~₹1,800/month. Add when revenue supports it |
| Deploy mechanism | SSH + docker compose on EC2 | Simplest for single-server. Upgrade to ECS when scaling |
| Frontend deploy | Vercel Git integration | Automatic on push to `main` — zero config |

---

## 2. Production Architecture

```
                    ┌─────────────────────────────────────┐
                    │         Hostinger DNS                │
                    │  careerdock.skriptvalley.com         │
                    │  api.careerdock.skriptvalley.com     │
                    │  assets.careerdock.skriptvalley.com  │
                    └───┬────────────┬──────────────┬──────┘
                        │            │              │
                   CNAME │       A record       CNAME │
                        │            │              │
                   ┌────▼────┐  ┌───▼───────┐  ┌───▼──────────┐
                   │ Vercel  │  │ EC2       │  │ CloudFront   │
                   │ Next.js │  │ Elastic IP│  │ (logos CDN)  │
                   │ Frontend│  │           │  │              │
                   └────┬────┘  └───┬───────┘  └───┬──────────┘
                        │           │              │
                        │ REST API  │              │ S3 Origin
                        ▼           ▼              ▼
              ┌─────────────────────────────────────────────┐
              │  EC2 t3.medium (Docker Compose)              │
              │                                              │
              │  ┌──────────┐  ┌──────┐  ┌───────────────┐  │
              │  │ Nginx    │  │ API  │  │ Worker        │  │
              │  │ :443/:80 │──│:8080 │  │ (Asynq)       │  │
              │  │ (SSL)    │  │      │  │               │  │
              │  └──────────┘  └──┬───┘  └───────┬───────┘  │
              │                   │              │           │
              │              ┌────▼──────────────▼────┐     │
              │              │ Redis :6379             │     │
              │              │ (sessions, cache, jobs) │     │
              │              └────────────────────────┘     │
              └───────────────────┬─────────────────────────┘
                                  │ VPC internal
                      ┌───────────▼───────────────┐
                      │ RDS PostgreSQL db.t3.micro │
                      │ (encrypted, no public IP)  │
                      └───────────────────────────┘

              ┌───────────────────────────────────┐
              │ S3 Buckets                         │
              │  careerdock-resumes (private)       │
              │  careerdock-logos (CloudFront OAC)  │
              └───────────────────────────────────┘

              ┌───────────────────────────────────┐
              │ ECR (Docker Registry)              │
              │  careerdock-api                     │
              │  careerdock-worker                  │
              └───────────────────────────────────┘
```

---

## 3. Infrastructure Setup

### 3.1 EC2 Instance

**One-time setup via AWS Console:**

| Setting | Value |
|---------|-------|
| AMI | Amazon Linux 2023 (latest) |
| Instance type | t3.medium (2 vCPU, 4GB RAM) |
| Storage | 30GB gp3 EBS |
| Elastic IP | Allocated and associated |
| Key pair | `careerdock-prod` (Ed25519, stored securely) |
| IAM instance profile | `careerdock-ec2-role` (see §3.1.2) |

**Security group: `careerdock-api-sg`**

| Direction | Port | Source | Purpose |
|-----------|------|--------|---------|
| Inbound | 443 | 0.0.0.0/0 | HTTPS (Nginx) |
| Inbound | 80 | 0.0.0.0/0 | HTTP → 301 redirect to HTTPS |
| Inbound | 22 | `<admin-ip>/32` | SSH (restricted) |
| Outbound | All | 0.0.0.0/0 | API calls, package updates |

#### 3.1.1 EC2 Bootstrap Script

Run once after instance creation:

```bash
#!/bin/bash
# 1. Update system
sudo dnf update -y

# 2. Install Docker
sudo dnf install -y docker
sudo systemctl enable docker
sudo systemctl start docker

# 3. Install Docker Compose v2
sudo mkdir -p /usr/local/lib/docker/cli-plugins
sudo curl -SL https://github.com/docker/compose/releases/latest/download/docker-compose-linux-x86_64 \
  -o /usr/local/lib/docker/cli-plugins/docker-compose
sudo chmod +x /usr/local/lib/docker/cli-plugins/docker-compose

# 4. Add ec2-user to docker group
sudo usermod -aG docker ec2-user

# 5. Install certbot (Let's Encrypt)
sudo dnf install -y certbot

# 6. Install AWS CLI (for ECR login, Secrets Manager)
# Pre-installed on Amazon Linux 2023

# 7. Create app directory
sudo mkdir -p /opt/careerdock
sudo chown ec2-user:ec2-user /opt/careerdock

# 8. Enable unattended security updates
sudo dnf install -y dnf-automatic
sudo systemctl enable dnf-automatic-install.timer
sudo systemctl start dnf-automatic-install.timer
```

#### 3.1.2 IAM Instance Profile — `careerdock-ec2-role`

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:BatchGetImage",
        "ecr:GetDownloadUrlForLayer"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:PutObject",
        "s3:GetObject",
        "s3:DeleteObject"
      ],
      "Resource": [
        "arn:aws:s3:::careerdock-resumes/*",
        "arn:aws:s3:::careerdock-logos/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:GetSecretValue"
      ],
      "Resource": "arn:aws:secretsmanager:ap-south-1:*:secret:careerdock/*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:ap-south-1:*:log-group:careerdock-*"
    }
  ]
}
```

### 3.2 RDS PostgreSQL

**Setup via AWS Console:**

| Setting | Value |
|---------|-------|
| Engine | PostgreSQL 16 |
| Instance class | db.t3.micro (free tier eligible Year 1) |
| Storage | 20GB gp3, autoscaling up to 50GB |
| Multi-AZ | Disabled (MVP — enable when revenue supports) |
| Public access | No |
| VPC | Default VPC (same as EC2) |
| Security group | `careerdock-rds-sg` |
| Encryption at rest | Enabled (AWS-managed KMS key) |
| Automated backups | 7-day retention |
| Deletion protection | Enabled |
| Database name | `careerdock` |
| Master username | `careerdock_admin` |
| Parameter group | Default (customise in staging for debug logging) |

**Security group: `careerdock-rds-sg`**

| Direction | Port | Source | Purpose |
|-----------|------|--------|---------|
| Inbound | 5432 | `careerdock-api-sg` | EC2 → RDS only |

**Connection string:**
```
postgres://careerdock_admin:<password>@careerdock-db.xxxxxx.ap-south-1.rds.amazonaws.com:5432/careerdock?sslmode=require
```

### 3.3 S3 Buckets

#### `careerdock-resumes` (Private)

| Setting | Value |
|---------|-------|
| Region | ap-south-1 (Mumbai) |
| Versioning | Disabled (PDFs are immutable — identified by resume_id) |
| Block public access | All 4 toggles ON |
| Encryption | SSE-S3 (AES-256, AWS-managed) |
| Lifecycle rule | Delete objects 90 days after creation with prefix `deleted/` |

Resume deletion flow: when a user's account is hard-deleted (30 days post soft-delete), the worker moves PDFs to `deleted/{user_id}/{resume_id}.pdf` prefix. The lifecycle rule auto-purges them after 90 additional days.

#### `careerdock-logos` (Public via CloudFront)

| Setting | Value |
|---------|-------|
| Region | ap-south-1 |
| Versioning | Disabled |
| Block public access | All 4 toggles ON (access via CloudFront OAC only) |
| Encryption | SSE-S3 |

### 3.4 CloudFront Distribution

| Setting | Value |
|---------|-------|
| Origin | S3 `careerdock-logos` |
| Origin access | OAC (Origin Access Control) — S3 bucket policy grants read to CloudFront only |
| Alternate domain | `assets.careerdock.skriptvalley.com` |
| SSL certificate | AWS Certificate Manager (free, for CloudFront custom domain) |
| Cache policy | CachingOptimized (TTL: default 24h, max 31d) |
| Price class | PriceClass_200 (US, Europe, Asia — excludes South America, Australia) |
| Compress objects | Yes (gzip + brotli) |

**S3 bucket policy for CloudFront OAC:**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "cloudfront.amazonaws.com"
      },
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::careerdock-logos/*",
      "Condition": {
        "StringEquals": {
          "AWS:SourceArn": "arn:aws:cloudfront::<account-id>:distribution/<distribution-id>"
        }
      }
    }
  ]
}
```

### 3.5 ECR Repositories

Create two private repositories:

```bash
aws ecr create-repository --repository-name careerdock-api --region ap-south-1
aws ecr create-repository --repository-name careerdock-worker --region ap-south-1
```

**Lifecycle policy** (keep last 10 images, auto-delete untagged after 1 day):
```json
{
  "rules": [
    {
      "rulePriority": 1,
      "description": "Remove untagged images after 1 day",
      "selection": {
        "tagStatus": "untagged",
        "countType": "sinceImagePushed",
        "countUnit": "days",
        "countNumber": 1
      },
      "action": { "type": "expire" }
    },
    {
      "rulePriority": 2,
      "description": "Keep last 10 tagged images",
      "selection": {
        "tagStatus": "tagged",
        "tagPatternList": ["v*"],
        "countType": "imageCountMoreThan",
        "countNumber": 10
      },
      "action": { "type": "expire" }
    }
  ]
}
```

### 3.6 DNS Records (Hostinger)

| Record | Type | Name | Value | TTL |
|--------|------|------|-------|-----|
| Frontend | CNAME | `careerdock` | `cname.vercel-dns.com` | 3600 |
| API | A | `api.careerdock` | `<EC2-Elastic-IP>` | 300 |
| Assets CDN | CNAME | `assets.careerdock` | `<cloudfront-distribution>.cloudfront.net` | 3600 |

### 3.7 SSL Certificate — Let's Encrypt

**Initial setup on EC2:**

```bash
# Stop Nginx temporarily (certbot needs port 80)
docker compose -f /opt/careerdock/docker-compose.prod.yml stop nginx

# Obtain certificate
sudo certbot certonly --standalone \
  -d api.careerdock.skriptvalley.com \
  --email admin@skriptvalley.com \
  --agree-tos \
  --non-interactive

# Restart Nginx
docker compose -f /opt/careerdock/docker-compose.prod.yml start nginx
```

**Auto-renewal cron (runs twice daily, only renews if <30 days to expiry):**

```bash
# /etc/cron.d/certbot-renew
0 3,15 * * * root certbot renew --pre-hook "docker compose -f /opt/careerdock/docker-compose.prod.yml stop nginx" --post-hook "docker compose -f /opt/careerdock/docker-compose.prod.yml start nginx" >> /var/log/certbot-renew.log 2>&1
```

---

## 4. Production Docker Compose

```yaml
# /opt/careerdock/docker-compose.prod.yml

services:
  api:
    image: ${ECR_REGISTRY}/careerdock-api:${VERSION}
    restart: unless-stopped
    env_file: .env.prod
    ports:
      - "127.0.0.1:8080:8080"   # Internal only — Nginx proxies
    depends_on:
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8080/api/health"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s
    deploy:
      resources:
        limits:
          memory: 1G
          cpus: '1.0'
    logging:
      driver: awslogs
      options:
        awslogs-group: careerdock-api
        awslogs-region: ap-south-1
        awslogs-stream-prefix: api

  worker:
    image: ${ECR_REGISTRY}/careerdock-worker:${VERSION}
    restart: unless-stopped
    env_file: .env.prod
    depends_on:
      redis:
        condition: service_healthy
    deploy:
      resources:
        limits:
          memory: 1G
          cpus: '1.0'
    logging:
      driver: awslogs
      options:
        awslogs-group: careerdock-worker
        awslogs-region: ap-south-1
        awslogs-stream-prefix: worker

  redis:
    image: redis:7-alpine
    restart: unless-stopped
    command: >
      redis-server
        --appendonly yes
        --maxmemory 256mb
        --maxmemory-policy allkeys-lru
        --requirepass ${REDIS_PASSWORD}
    ports:
      - "127.0.0.1:6379:6379"   # Internal only
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "${REDIS_PASSWORD}", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5

  nginx:
    image: nginx:1.25-alpine
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - /etc/letsencrypt:/etc/letsencrypt:ro
    depends_on:
      api:
        condition: service_healthy

  migrate:
    image: ${ECR_REGISTRY}/careerdock-api:${VERSION}
    command: ["/usr/local/bin/api", "migrate", "up"]
    env_file: .env.prod
    profiles: ["migration"]                      # Only runs when explicitly invoked
    depends_on:
      redis:
        condition: service_healthy

volumes:
  redis_data:
```

### 4.1 Production Environment File

```bash
# /opt/careerdock/.env.prod — loaded from AWS Secrets Manager by deploy script
# This file is generated on each deploy, never committed

PORT=8080
ENVIRONMENT=production
DATABASE_URL=postgres://careerdock_admin:<password>@<rds-endpoint>:5432/careerdock?sslmode=require
REDIS_URL=redis://:${REDIS_PASSWORD}@redis:6379
JWT_SECRET=<from-secrets-manager>
CLAUDE_API_KEY=<from-secrets-manager>
OPENAI_API_KEY=<from-secrets-manager>
RAZORPAY_KEY_ID=<from-secrets-manager>
RAZORPAY_KEY_SECRET=<from-secrets-manager>
RAZORPAY_WEBHOOK_SECRET=<from-secrets-manager>
RESEND_API_KEY=<from-secrets-manager>
FROM_EMAIL=noreply@careerdock.skriptvalley.com
S3_ENDPOINT=https://s3.ap-south-1.amazonaws.com
S3_REGION=ap-south-1
S3_RESUME_BUCKET=careerdock-resumes
S3_LOGO_BUCKET=careerdock-logos
S3_USE_PATH_STYLE=false
ALLOWED_ORIGINS=https://careerdock.skriptvalley.com
SENTRY_DSN=<from-secrets-manager>
```

Note: S3 credentials are not in `.env.prod` — the EC2 instance profile (IAM role) provides them automatically via the AWS SDK credential chain.

---

## 5. CI/CD Pipeline

### 5.1 Pipeline Overview

```
Developer pushes tag
        │
        ▼
┌──────────────────┐
│ GitHub Actions    │
│ deploy.yml        │
│                  │
│ 1. Build images  │
│ 2. Push to ECR   │
│ 3. SSH to EC2    │
│ 4. Pull images   │
│ 5. Run migrations│
│ 6. Restart services│
│ 7. Health check  │
│ 8. Notify        │
└──────────────────┘
```

### 5.2 GitHub Repository Secrets

Configure in GitHub → Settings → Secrets → Actions:

| Secret | Value |
|--------|-------|
| `AWS_ACCESS_KEY_ID` | Deploy IAM user access key |
| `AWS_SECRET_ACCESS_KEY` | Deploy IAM user secret key |
| `AWS_REGION` | `ap-south-1` |
| `ECR_REGISTRY` | `<account-id>.dkr.ecr.ap-south-1.amazonaws.com` |
| `EC2_HOST` | EC2 Elastic IP address |
| `EC2_SSH_KEY` | Private key for `careerdock-prod` key pair |
| `EC2_USER` | `ec2-user` |

**Deploy IAM user policy** (separate from EC2 instance profile):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:BatchCheckLayerAvailability",
        "ecr:InitiateLayerUpload",
        "ecr:UploadLayerPart",
        "ecr:CompleteLayerUpload",
        "ecr:PutImage"
      ],
      "Resource": "*"
    }
  ]
}
```

### 5.3 Deploy Workflow

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    tags:
      - 'v*'

env:
  AWS_REGION: ap-south-1
  ECR_REGISTRY: ${{ secrets.ECR_REGISTRY }}

jobs:
  build-and-push:
    name: Build & Push Docker Images
    runs-on: ubuntu-latest
    outputs:
      version: ${{ steps.version.outputs.tag }}

    steps:
      - uses: actions/checkout@v4

      - name: Extract version from tag
        id: version
        run: echo "tag=${GITHUB_REF#refs/tags/}" >> $GITHUB_OUTPUT

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: ${{ env.AWS_REGION }}

      - name: Login to ECR
        uses: aws-actions/amazon-ecr-login@v2

      - name: Build and push API image
        working-directory: backend
        run: |
          docker build -f ../infra/docker/Dockerfile.api \
            -t $ECR_REGISTRY/careerdock-api:${{ steps.version.outputs.tag }} \
            -t $ECR_REGISTRY/careerdock-api:latest \
            .
          docker push $ECR_REGISTRY/careerdock-api:${{ steps.version.outputs.tag }}
          docker push $ECR_REGISTRY/careerdock-api:latest

      - name: Build and push Worker image
        working-directory: backend
        run: |
          docker build -f ../infra/docker/Dockerfile.worker \
            -t $ECR_REGISTRY/careerdock-worker:${{ steps.version.outputs.tag }} \
            -t $ECR_REGISTRY/careerdock-worker:latest \
            .
          docker push $ECR_REGISTRY/careerdock-worker:${{ steps.version.outputs.tag }}
          docker push $ECR_REGISTRY/careerdock-worker:latest

  deploy:
    name: Deploy to Production
    needs: build-and-push
    runs-on: ubuntu-latest
    environment: production                       # Requires manual approval

    steps:
      - uses: actions/checkout@v4

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: ${{ env.AWS_REGION }}

      - name: Deploy to EC2
        uses: appleboy/ssh-action@v1
        with:
          host: ${{ secrets.EC2_HOST }}
          username: ${{ secrets.EC2_USER }}
          key: ${{ secrets.EC2_SSH_KEY }}
          script: |
            set -e
            cd /opt/careerdock

            # 1. Login to ECR
            aws ecr get-login-password --region ap-south-1 | \
              docker login --username AWS --password-stdin ${{ secrets.ECR_REGISTRY }}

            # 2. Set version
            export VERSION="${{ needs.build-and-push.outputs.version }}"
            export ECR_REGISTRY="${{ secrets.ECR_REGISTRY }}"

            # 3. Pull new images
            docker compose -f docker-compose.prod.yml pull api worker

            # 4. Fetch secrets from AWS Secrets Manager and write .env.prod
            /opt/careerdock/scripts/load-secrets.sh

            # 5. Run migrations
            docker compose -f docker-compose.prod.yml run --rm migrate

            # 6. Rolling restart (zero-downtime for API)
            docker compose -f docker-compose.prod.yml up -d --no-deps api worker

            # 7. Health check (wait up to 60s)
            for i in $(seq 1 12); do
              if curl -sf http://localhost:8080/api/health > /dev/null; then
                echo "Health check passed"
                exit 0
              fi
              echo "Waiting for health check... ($i/12)"
              sleep 5
            done

            echo "Health check failed — rolling back"
            export VERSION="${{ needs.build-and-push.outputs.version }}-rollback"
            # Rollback: restart with previous images
            docker compose -f docker-compose.prod.yml up -d --no-deps api worker
            exit 1
```

### 5.4 Secrets Loader Script

```bash
#!/bin/bash
# /opt/careerdock/scripts/load-secrets.sh
# Fetches secrets from AWS Secrets Manager and writes .env.prod

set -e

SECRET_ID="careerdock/production"
REGION="ap-south-1"

# Fetch all secrets as JSON
SECRETS=$(aws secretsmanager get-secret-value \
  --secret-id "$SECRET_ID" \
  --region "$REGION" \
  --query SecretString \
  --output text)

# Write .env.prod
cat > /opt/careerdock/.env.prod << EOF
PORT=8080
ENVIRONMENT=production
DATABASE_URL=$(echo "$SECRETS" | jq -r '.DATABASE_URL')
REDIS_URL=redis://:$(echo "$SECRETS" | jq -r '.REDIS_PASSWORD')@redis:6379
JWT_SECRET=$(echo "$SECRETS" | jq -r '.JWT_SECRET')
CLAUDE_API_KEY=$(echo "$SECRETS" | jq -r '.CLAUDE_API_KEY')
OPENAI_API_KEY=$(echo "$SECRETS" | jq -r '.OPENAI_API_KEY')
RAZORPAY_KEY_ID=$(echo "$SECRETS" | jq -r '.RAZORPAY_KEY_ID')
RAZORPAY_KEY_SECRET=$(echo "$SECRETS" | jq -r '.RAZORPAY_KEY_SECRET')
RAZORPAY_WEBHOOK_SECRET=$(echo "$SECRETS" | jq -r '.RAZORPAY_WEBHOOK_SECRET')
RESEND_API_KEY=$(echo "$SECRETS" | jq -r '.RESEND_API_KEY')
FROM_EMAIL=noreply@careerdock.skriptvalley.com
S3_ENDPOINT=https://s3.ap-south-1.amazonaws.com
S3_REGION=ap-south-1
S3_RESUME_BUCKET=careerdock-resumes
S3_LOGO_BUCKET=careerdock-logos
S3_USE_PATH_STYLE=false
ALLOWED_ORIGINS=https://careerdock.skriptvalley.com
SENTRY_DSN=$(echo "$SECRETS" | jq -r '.SENTRY_DSN')
REDIS_PASSWORD=$(echo "$SECRETS" | jq -r '.REDIS_PASSWORD')
EOF

chmod 600 /opt/careerdock/.env.prod
echo "Secrets loaded into .env.prod"
```

### 5.5 Frontend Deployment (Vercel)

| Setting | Value |
|---------|-------|
| Platform | Vercel (free tier) |
| Git integration | Connected to `main` branch |
| Build command | `npm run build` (automatic) |
| Deploy trigger | Every push to `main` |
| Preview deploys | On PRs (automatic) |
| Environment variables | `NEXT_PUBLIC_API_URL=https://api.careerdock.skriptvalley.com` |
| Custom domain | `careerdock.skriptvalley.com` |

Vercel deployment is fully automatic — no GitHub Actions workflow needed. Push to `main` → Vercel builds → deploys.

---

## 6. Deployment Procedures

### 6.1 Standard Release

```bash
# 1. Ensure main is clean and CI passes
git checkout main
git pull origin main

# 2. Tag the release
git tag -a v1.2.0 -m "Release v1.2.0: Add job ATS check feature"
git push origin v1.2.0

# 3. GitHub Actions:
#    - build-and-push job runs automatically
#    - deploy job waits for manual approval (GitHub "production" environment)

# 4. Approve deploy in GitHub Actions UI

# 5. Monitor:
#    - GitHub Actions logs for deploy status
#    - /api/health for service health
#    - Sentry for new errors
#    - CloudWatch logs for API/worker output
```

### 6.2 Hotfix Release

```bash
# 1. Branch from the broken release tag
git checkout -b hotfix/v1.2.x v1.2.0

# 2. Apply fix
git cherry-pick <fix-commit-sha>
# or make the fix directly

# 3. Tag the patch
git tag -a v1.2.1 -m "Hotfix: Fix credit allocation race condition"
git push origin v1.2.1

# 4. Merge fix back to main
git checkout main
git merge hotfix/v1.2.x
git push origin main

# 5. Delete hotfix branch
git branch -d hotfix/v1.2.x
git push origin --delete hotfix/v1.2.x

# 6. Deploy follows standard pipeline (tag triggers GitHub Actions)
```

### 6.3 Database Migration Deploy

Migrations run automatically as part of the deploy pipeline (step 5 in §5.3). For manual migration operations:

```bash
# SSH into EC2
ssh -i careerdock-prod.pem ec2-user@<elastic-ip>

# Run pending migrations
cd /opt/careerdock
docker compose -f docker-compose.prod.yml run --rm migrate

# Check current migration version
docker compose -f docker-compose.prod.yml run --rm migrate version

# Rollback last migration (use with caution)
docker compose -f docker-compose.prod.yml run --rm migrate down 1
```

### 6.4 Rollback Procedure

**Scenario:** v1.3.0 deployed, bugs found, need to revert to v1.2.1.

```bash
# SSH into EC2
ssh -i careerdock-prod.pem ec2-user@<elastic-ip>
cd /opt/careerdock

# 1. Set version to last known-good
export VERSION=v1.2.1
export ECR_REGISTRY=<account-id>.dkr.ecr.ap-south-1.amazonaws.com

# 2. Pull the old images (should be in ECR cache)
docker compose -f docker-compose.prod.yml pull api worker

# 3. Restart with old version
docker compose -f docker-compose.prod.yml up -d --no-deps api worker

# 4. If v1.3.0 included a migration, roll it back:
# WARNING: Only do this if the migration is reversible and no data depends on it
docker compose -f docker-compose.prod.yml run --rm \
  -e VERSION=v1.3.0 migrate down 1

# 5. Verify health
curl -sf http://localhost:8080/api/health
```

**Database rollback risks:**
- Data written by the new migration schema may be lost or corrupted.
- Always prefer forward-fixing (new migration + new deploy) over rollback.
- Only roll back migrations if no data has been written to new schema objects.

---

## 7. Health Checks

### 7.1 API Health Endpoint

`GET /api/health` — no authentication required.

```go
// internal/handler/health.go

type HealthResponse struct {
    Status    string            `json:"status"`
    Version   string            `json:"version"`
    Timestamp time.Time         `json:"timestamp"`
    Checks    map[string]string `json:"checks"`
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
    checks := map[string]string{}
    allOK := true

    // Database
    if err := db.Ping(r.Context()); err != nil {
        checks["database"] = "unhealthy"
        allOK = false
    } else {
        checks["database"] = "ok"
    }

    // Redis
    if err := redis.Ping(r.Context()).Err(); err != nil {
        checks["redis"] = "unhealthy"
        allOK = false
    } else {
        checks["redis"] = "ok"
    }

    status := "ok"
    httpStatus := http.StatusOK
    if !allOK {
        status = "degraded"
        httpStatus = http.StatusServiceUnavailable
    }

    respondJSON(w, httpStatus, HealthResponse{
        Status:    status,
        Version:   buildVersion, // Injected at build time via -ldflags
        Timestamp: time.Now(),
        Checks:    checks,
    })
}
```

### 7.2 Build Version Injection

```dockerfile
# In Dockerfile.api
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-X main.buildVersion=${VERSION}" -o /api ./cmd/api/
```

```yaml
# In deploy.yml build step
docker build --build-arg VERSION=${{ steps.version.outputs.tag }} ...
```

This makes the health endpoint return the exact git tag running in production:
```json
{
  "status": "ok",
  "version": "v1.2.0",
  "timestamp": "2026-03-12T10:00:00Z",
  "checks": { "database": "ok", "redis": "ok" }
}
```

### 7.3 External Monitoring

| Tool | What It Monitors | Frequency | Alert |
|------|-----------------|-----------|-------|
| UptimeRobot (free) | `GET https://api.careerdock.skriptvalley.com/api/health` | Every 5 min | Email + push notification |
| UptimeRobot (free) | `GET https://careerdock.skriptvalley.com` | Every 5 min | Email |
| Sentry | Runtime errors (API + worker + frontend) | Real-time | Email on new issue |

---

## 8. Logging

### 8.1 Log Destinations

| Component | Log Driver | Destination |
|-----------|-----------|-------------|
| API | `awslogs` | CloudWatch `careerdock-api` log group |
| Worker | `awslogs` | CloudWatch `careerdock-worker` log group |
| Nginx | `local` | `/var/log/nginx/access.log`, `/var/log/nginx/error.log` on EC2 |
| Redis | Docker default | `docker compose logs redis` |

### 8.2 CloudWatch Log Groups

Create before first deploy:

```bash
aws logs create-log-group --log-group-name careerdock-api --region ap-south-1
aws logs create-log-group --log-group-name careerdock-worker --region ap-south-1

# Set retention (30 days for MVP — keeps costs low)
aws logs put-retention-policy --log-group-name careerdock-api --retention-in-days 30
aws logs put-retention-policy --log-group-name careerdock-worker --retention-in-days 30
```

### 8.3 Log Format

All Go logs use `slog` with JSON output:

```json
{
  "time": "2026-03-12T10:05:32.123Z",
  "level": "INFO",
  "msg": "request completed",
  "request_id": "abc-123",
  "method": "GET",
  "path": "/api/companies",
  "status": 200,
  "duration_ms": 12,
  "user_id": "01912345-6789-7abc-..."
}
```

### 8.4 Viewing Logs

```bash
# Tail API logs (real-time)
ssh ec2-user@<ip> "docker compose -f /opt/careerdock/docker-compose.prod.yml logs -f api"

# Via CloudWatch CLI
aws logs tail careerdock-api --region ap-south-1 --follow

# Search for errors
aws logs filter-log-events \
  --log-group-name careerdock-api \
  --filter-pattern '"level":"ERROR"' \
  --region ap-south-1
```

---

## 9. Operational Runbooks

### 9.1 Database Backup & Restore

**Backups are automatic** (RDS, 7-day retention). For manual snapshot:

```bash
# Create manual snapshot
aws rds create-db-snapshot \
  --db-instance-identifier careerdock-db \
  --db-snapshot-identifier careerdock-pre-migration-$(date +%Y%m%d) \
  --region ap-south-1

# Restore from snapshot (creates new RDS instance)
aws rds restore-db-instance-from-db-snapshot \
  --db-instance-identifier careerdock-db-restored \
  --db-snapshot-identifier careerdock-pre-migration-20260312 \
  --region ap-south-1
```

**Always create a manual snapshot before risky migrations.**

### 9.2 Redis Data Recovery

Redis uses AOF persistence. If Redis container crashes:

```bash
# Redis data is in the named volume — just restart
docker compose -f docker-compose.prod.yml restart redis

# If volume is corrupted, Redis starts empty (sessions/cache lost, not critical):
# - Users will be logged out (refresh tokens lost) — they re-login
# - AI cache will rebuild on next request
# - Rate limit counters reset
# - Asynq job queue state is lost — check for stuck jobs in admin dashboard
```

### 9.3 Emergency Credential Rotation

```bash
# 1. Update secret in AWS Secrets Manager
aws secretsmanager update-secret \
  --secret-id careerdock/production \
  --secret-string '{"JWT_SECRET":"new-value",...}' \
  --region ap-south-1

# 2. Reload secrets on EC2
ssh ec2-user@<ip> "/opt/careerdock/scripts/load-secrets.sh"

# 3. Restart services to pick up new env
ssh ec2-user@<ip> "cd /opt/careerdock && docker compose -f docker-compose.prod.yml restart api worker"
```

For JWT secret rotation specifically, use the dual-key window described in [SECURITY.md §6.4](./SECURITY.md).

### 9.4 Disk Space Issues

```bash
# Check disk usage
df -h

# Docker cleanup (remove unused images, containers, networks)
docker system prune -f

# Remove old Docker images (keep last 3 versions)
docker images --format '{{.Repository}}:{{.Tag}}' | grep careerdock | sort -V | head -n -6 | xargs -r docker rmi

# Check CloudWatch log size
du -sh /var/lib/docker/containers/*/
```

### 9.5 Scaling Up (When Needed)

| Trigger | Action | Downtime |
|---------|--------|----------|
| CPU consistently >80% | Upgrade EC2 to t3.large | ~5 min (stop/start) |
| Memory consistently >80% | Upgrade EC2 to t3.large | ~5 min |
| DB connections exhausted | Upgrade RDS to db.t3.small | ~10 min (RDS apply) |
| Redis memory full | Increase maxmemory or move to ElastiCache | Minutes (config change) |
| S3 costs growing | Add lifecycle rules for old data | None |

---

## 10. Cost Summary

### 10.1 Year 1 (Free Tier Eligible)

| Service | Monthly Cost (₹) | Notes |
|---------|------------------:|-------|
| EC2 t3.medium (on-demand) | 2,500 | Can reduce to ~1,500 with Reserved Instance |
| RDS db.t3.micro | 0 | Free tier (750 hrs/month for 12 months) |
| S3 | ~10 | 5GB free, minimal usage at MVP |
| CloudFront | 0 | 1TB/month free |
| ECR | 0 | 500MB free storage |
| CloudWatch Logs | 0 | 5GB ingestion/month free |
| Vercel | 0 | Free tier |
| Resend | 0 | 3K emails/month free |
| Sentry | 0 | 5K events/month free |
| UptimeRobot | 0 | 50 monitors free |
| DNS (Hostinger) | 0 | Included with domain |
| SSL (Let's Encrypt) | 0 | Free |
| **Total** | **~₹2,510** | |

### 10.2 Year 2+ (Post Free Tier)

| Service | Monthly Cost (₹) | Notes |
|---------|------------------:|-------|
| EC2 t3.medium (Reserved 1yr) | ~1,500 | ~40% savings over on-demand |
| RDS db.t3.micro | 1,200 | Post-free-tier pricing |
| S3 + CloudFront + ECR | ~50 | Minimal at MVP scale |
| CloudWatch Logs | ~100 | 30-day retention |
| **Total** | **~₹2,850** | |

### 10.3 AI Costs (Variable)

| Metric | Estimate |
|--------|----------|
| AI cost per Starter Pack consumed | ~₹40 |
| Starter Pack price | ₹399 |
| **Gross margin on AI** | **~90%** |

AI costs scale linearly with usage — no fixed infrastructure cost.

---

## 11. First Deploy Checklist

Complete in order:

### AWS Resources
- [ ] Create VPC (or use default) in `ap-south-1`
- [ ] Create EC2 security group (`careerdock-api-sg`)
- [ ] Launch EC2 t3.medium with Amazon Linux 2023
- [ ] Allocate and associate Elastic IP
- [ ] Run bootstrap script (§3.1.1)
- [ ] Create RDS PostgreSQL instance
- [ ] Create RDS security group (`careerdock-rds-sg`)
- [ ] Create S3 buckets (`careerdock-resumes`, `careerdock-logos`)
- [ ] Create CloudFront distribution for logos
- [ ] Create ECR repositories (`careerdock-api`, `careerdock-worker`)
- [ ] Create IAM instance profile (`careerdock-ec2-role`)
- [ ] Create IAM deploy user for GitHub Actions
- [ ] Create AWS Secrets Manager secret (`careerdock/production`)
- [ ] Create CloudWatch log groups

### DNS & SSL
- [ ] Add DNS records in Hostinger (§3.6)
- [ ] Obtain Let's Encrypt certificate (§3.7)
- [ ] Set up auto-renewal cron

### GitHub
- [ ] Add repository secrets (§5.2)
- [ ] Create `production` environment with required reviewers
- [ ] Push `deploy.yml` workflow

### Deploy
- [ ] Copy `docker-compose.prod.yml` and `nginx.conf` to EC2
- [ ] Copy `load-secrets.sh` to EC2
- [ ] Run `load-secrets.sh` to generate `.env.prod`
- [ ] Tag first release: `git tag -a v1.0.0 -m "Initial release"`
- [ ] Push tag and approve deploy
- [ ] Verify `/api/health` returns `ok`

### Vercel
- [ ] Connect GitHub repo to Vercel
- [ ] Set `NEXT_PUBLIC_API_URL` environment variable
- [ ] Add custom domain `careerdock.skriptvalley.com`
- [ ] Verify frontend loads

### Monitoring
- [ ] Add UptimeRobot monitors for API and frontend
- [ ] Configure Sentry project (backend + frontend)
- [ ] Verify CloudWatch logs are flowing

---

## 12. Staging Environment (Future)

When revenue supports it (~₹1,800/month additional), set up staging:

| Resource | Staging Value |
|----------|--------------|
| EC2 | t3.small (1 vCPU, 2GB) |
| RDS | db.t3.micro |
| S3 | `careerdock-resumes-staging`, `careerdock-logos-staging` |
| DNS | `staging-api.careerdock.skriptvalley.com` |
| SSL | Separate Let's Encrypt cert |
| Secrets | `careerdock/staging` in Secrets Manager |
| Deploy | Manual trigger, same workflow with staging environment |

Staging mirrors production topology at smaller scale. Feature flags can be tested in staging before enabling in production.

---

## 13. Scaling Roadmap

| Stage | Users | Infrastructure | Estimated Cost |
|-------|-------|---------------|---------------|
| **MVP** | 0 – 1K | Single EC2 t3.medium, RDS db.t3.micro, co-located Redis | ~₹2,500/month |
| **Growth** | 1K – 10K | EC2 t3.large, RDS db.t3.small, ElastiCache Redis | ~₹6,000/month |
| **Scale** | 10K+ | ECS Fargate (API), separate worker EC2, RDS read replica, MeiliSearch | ~₹15,000/month |

**Migration path is smooth** because:
1. Backend is already containerised — ECS Fargate is a config change, not a code change.
2. Redis is accessed via URL — swapping local Redis for ElastiCache is an env var change.
3. Database is already RDS — upgrading instance type or adding read replica requires no code changes.
4. Feature flags are database-backed — no infrastructure dependency.
