# CI/CD Guide

This document explains how to run the GitLab CI/CD pipeline for hiTech, configure self-hosted runners on target servers, and maintain deployment environments.

## 1. Pipeline Overview

The top-level pipeline (`.gitlab-ci.yml`) defines four stages:

| Stage | Jobs | Description |
| --- | --- | --- |
| `build` | `build:go-orchestrator`, `build:drone-service`, `build:admin-panel` | Builds Docker images, pushes to GitLab Container Registry. On `main` branch also updates `:latest` tags. |
| `test` | `test:go`, `test:drone-service`, `test:admin-panel` | Runs Go unit tests, Python pytest, and TypeScript lint. |
| `deploy` | `deploy:dev` (dev branch), `deploy:prod` (main branch) | SSH into servers, pulls images, executes `docker compose up -d --remove-orphans` using `deployment/docker/docker-compose.cicd.yml`. |
| `cleanup` | `docker:gc` | Optional image cache pruning on runners. |

### Trigger Flags
- `FORCE_BUILD_ALL=true`: rebuild all services regardless of changes.
- `FORCE_DEPLOY=true`: run deploy job even when tracked files have no changes.

Set these as CI/CD variables or provide when running manual pipeline (`Run pipeline` > Variables).

## 2. Required GitLab CI/CD Variables

Configure the following variables in the GitLab project (`Settings → CI/CD → Variables`). Mark secrets as **protected**/**masked** as appropriate.

### Registry & Auth
| Variable | Example | Purpose |
| --- | --- | --- |
| `CI_REGISTRY` | `registry.gitlab.com` | Provided automatically; ensure accessible from runners. |
| `CI_REGISTRY_USER` | `gitlab-ci-token` | Automatic for CI jobs. |
| `CI_REGISTRY_PASSWORD` | `${CI_JOB_TOKEN}` | Auto-populated. |

### SSH & Deployment Targets
| Variable | Example | Notes |
| --- | --- | --- |
| `SSH_PRIVATE_KEY` | (private key text) | Used to SSH into deployment servers; key must match `~/.ssh/authorized_keys`. |
| `DEV_SERVER_IP` | `192.168.10.20` | Dev host. |
| `DEV_SERVER_USER` | `deploy` | SSH user for dev. |
| `DEV_DEPLOY_PATH` | `/opt/hitech-dev` | Optional; defaults to `/opt/hitech-dev`. |
| `DEV_ENV_FILE` | `.env.dev` | Optional; defaults to `.env.dev`, fallback `.env.prod` if missing. |
| `PROD_SERVER_IP` | `203.0.113.10` | Production host. |
| `PROD_SERVER_USER` | `deploy` | SSH user for prod. |
| `PROD_DEPLOY_PATH` | `/opt/hitech` | Optional; defaults to `/opt/hitech`. |
| `PROD_ENV_FILE` | `.env.prod` | Optional; defaults to `.env.prod`. |

### Application Secrets
Store all values referenced in `.env.prod` / `.env.dev`: JWT secrets, database credentials, MinIO keys, SMSAero credentials, Firebase path, Grafana credentials, etc. **Do not commit secrets to the repo.**

## 3. Preparing Target Servers

1. **Install Docker & Compose plugin**
   ```bash
   sudo apt-get update
   sudo apt-get install -y docker.io docker-compose-plugin
   sudo usermod -aG docker $USER
   ```
   Log out/in for group changes.

2. **Clone repository** to deployment path (e.g. `/opt/hitech`). Deploy jobs will `git reset --hard` to branch tip, so no manual edits inside this directory.

3. **Create environment file** (`.env.prod`, `.env.dev`) containing all variables expected by compose files.

4. **Firebase credentials**
   - Place service account JSON at path referenced by `FIREBASE_CREDENTIALS_PATH` (default `./secrets/firebase-service-account.json`).
   - Ensure permissions allow Docker to mount the file (`chmod 600`).

5. **Open required ports** on firewall: HTTP ingress (80/443 or custom Nginx ports), RabbitMQ UI (optional), Grafana/Prometheus/MinIO if exposed.

## 4. Installing GitLab Runner (Self-Hosted)

Each deployment server needs a GitLab Runner dedicated to `deploy` jobs.

```bash
curl -L https://packages.gitlab.com/install/repositories/runner/gitlab-runner/script.deb.sh | sudo bash
sudo apt-get install gitlab-runner

sudo gitlab-runner register
```

During registration:
- **URL:** `https://gitlab.com/` (or your GitLab instance).
- **Token:** from project → Settings → CI/CD → Runners → “Expand” → “Register a runner”.
- **Description:** `hitech-dev` or `hitech-prod`.
- **Tags:** use `dev` for the dev server, `prod` for production. Matches job tags in `gitlab-ci-*.yml`.
- **Executor:** `shell` (preferred) or `docker`. If using `shell`, ensure Docker CLI is available for compose commands.

Start the runner:
```bash
sudo gitlab-runner start
```

Verify in GitLab UI that the runner is online and tagged correctly.

## 5. Running Pipelines

- **Automatic:** pushes to `dev` trigger build/test and, if affected files changed, `deploy:dev`. Same for `main` and `deploy:prod`.
- **Manual:** `CI/CD → Pipelines → Run pipeline`. Provide branch and optional flags (`FORCE_DEPLOY`, `FORCE_BUILD_ALL`).
- **Monitoring:** view job logs to ensure `docker compose` finished successfully (`docker compose ... ps` output at end of deploy job).

## 6. Troubleshooting

| Symptom | Cause | Resolution |
| --- | --- | --- |
| Deploy job fails to SSH | Wrong IP/user or key not added to `authorized_keys`. | Update `DEV/PROD_SERVER_*` vars, ensure key presence. |
| Docker compose cannot find `.env` | Missing file on server or incorrect `ENV_FILE`. | Copy `.env` to deploy path, adjust variable. |
| Images missing in registry | Build stage failed or tags differ. | Re-run pipeline with `FORCE_BUILD_ALL=true`. |
| Permission denied for Firebase file | Volume path incorrect or lacks read perms. | Verify `FIREBASE_CREDENTIALS_PATH` and file permissions. |
| Runner offline | GitLab Runner service stopped or token invalid. | `sudo gitlab-runner verify`, restart service, re-register if needed. |

## 7. Manual Rollback

1. SSH to server.
2. `cd /opt/hitech` (or configured path).
3. `git checkout <previous commit>` (or stash) and set `IMAGE_TAG` to desired SHA.
4. `docker compose --env-file <env> -f deployment/docker/docker-compose.cicd.yml pull <services>`.
5. `docker compose ... up -d --remove-orphans`.

## 8. References
- `deployment/docker/docker-compose.dev.yml` / `prod.yml` for default environment settings.
- `docs/en/DEPLOYMENT.md` for detailed manual setup.
- `docs/en/OBSERVABILITY.md` for monitoring after deployment.
