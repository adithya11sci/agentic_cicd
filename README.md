# Agentic CI/CD Pipeline Repair & Intelligent Release Orchestration

[![Go Version](https://img.shields.io/badge/go-1.24-blue)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)

A production-grade "Self-Healing CI/CD Pipeline" built in Go. This system acts as an autonomous AI-driven DevOps engineer. It monitors CI/CD pipelines, diagnoses failures using contextual signals (commit diffs, logs, test reports) via LLMs, automatically generates code configurations/fixes, and enforces a human-in-the-loop governance layer before raising a final Pull Request to remediate the issue.

---

## Table of Contents
- [Architecture & Workflow](#architecture--workflow)
- [Multi-Agent System](#multi-agent-system)
- [Technology Stack](#technology-stack)
- [Prerequisites](#prerequisites)
- [Installation & Setup](#installation--setup)
- [Configuration](#configuration)
- [Usage & Testing](#usage--testing)
- [Project Structure](#project-structure)
- [Future Enhancements](#future-enhancements)

---

## Architecture & Workflow

The orchestration strictly follows an autonomous workflow driven by a centralized coordinator orchestrating discrete AI agents.

1. **Trigger:** A GitHub webhook fires on a `workflow_run` failure.
2. **Detection:** The Gin server catches the payload and alerts the **Monitor Agent**.
3. **Context Gathering:** The system fetches pipeline logs, test reports, and the breaking commit diff.
4. **Analysis:** The **Root Cause Agent** uses an LLM to interpret the failure in human terms.
5. **Remediation:** The **Repair Agent** writes a unified git patch intended to fix the failure.
6. **Governance:** The **Governance Agent** reviews the fix. If it's high-risk (e.g., core infrastructure), it flags for human approval. If low-risk (e.g., typo, missing dependency), it auto-approves.
7. **Resolution:** The **PR Agent** creates an automatic branch and raises a Pull Request with the fix and detailed explanations.

---

## Multi-Agent System

This system implements 5 specialized Agents inside a monolithic architecture pattern:

1. **Pipeline Monitoring Agent:** Acts as the entry point, listening to GitHub webhooks, validating payloads, and extracting core pipeline metadata.
2. **Root Cause Analysis Agent:** Consumes `Prompts + Context` to identify exact failure footprints (build error, dependency conflict, test failure).
3. **Auto Repair Agent:** Engineered to generate *minimal* structural patches to existing repositories based on identified root causes.
4. **Governance Agent:** Enforces security and risk policy rules (LOW/MEDIUM/HIGH) evaluating the fix stringency.
5. **Pull Request Agent:** Interfaces directly with the GitHub API to manage branches, commits, and PR descriptions for automated merge readiness.
6. **Orchestrator:** The main brain that asynchronously chains the output of one agent into the input of the next.

---

## Technology Stack

- **Go (Golang) 1.24:** Core application language.
- **Gin Framework:** Lightning-fast webhook HTTP server.
- **go-openai:** Wrapper to interact natively with OpenAI GPT-4.
- **go-github (v69):** To fetch diffs, logs, and interact with repository trees.
- **Uber Zap:** High-performance structured logging.
- **PostgreSQL:** (Prepared as a Docker service) for future data store of pipeline metrics.
- **Docker & Docker Compose:** Containerization & Rootless security profiles.

---

## Prerequisites

- [Go 1.24+](https://go.dev/doc/install) (If running locally instead of Docker)
- [Docker & Docker Compose](https://docs.docker.com/get-docker/)
- An **OpenAI API Key** with access to GPT-4 (or a compatible LLM).
- A **GitHub Personal Access Token (PAT)** with `repo` and `workflow` scopes.
- [ngrok](https://ngrok.com/) (Recommended for local webhook testing).

---

## Installation & Setup

1. **Clone the repository:**
   ```bash
   git clone https://github.com/adithya11sci/agentic_cicd.git
   cd agentic_cicd
   ```

2. **Environment Configuration:**
   Copy the example environment variables file and insert your keys.
   ```bash
   cp .env.example .env
   ```

3. **Running via Docker (Recommended):**
   We provide a fully dockerized setup using a secure, rootless profile.
   ```bash
   docker-compose up --build -d
   ```
   > The application will safely spin up on `localhost:8080`.

4. **Running Locally (Go Native):**
   ```bash
   go mod tidy
   go run cmd/server.go
   ```

---

## Configuration

The `.env` file exposes the following core configurations:

| Variable | Description | Example |
|----------|-------------|---------|
| `PORT` | Listening port for the Gin Webhook server | `8080` |
| `LLM_API_KEY` | Your OpenAI API key for LLM agents | `sk-...` |
| `GITHUB_TOKEN` | Token for reading diffs/logs & opening PRs | `ghp_...` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://user...`|

---

## Usage & Testing

### Simulating a Pipeline Failure Webhook

To securely expose your local instance to GitHub to receive live Webhooks:
```bash
ngrok http 8080
```

**In your GitHub Repository:**
1. Go to `Settings` -> `Webhooks` -> `Add webhook`.
2. **Payload URL**: `https://<your-ngrok-url>.ngrok.io/webhook/github`
3. **Content type**: `application/json`
4. **Events**: Select **Workflow runs**.
5. Save. GitHub will send a `ping` event. The Monitor agent gracefully acknowledges it.

Trigger a broken build manually in your repository, and observe the Go logs via `docker-compose logs -f` to see the agents in action!

---

## Project Structure

```text
agentic-cicd/
|
|-- cmd/
|   \-- server.go             # Application entrypoint & HTTP server
|
|-- internal/
|   |-- agents/               # Business logic for all 5 Agents
|   |   |-- monitor.go
|   |   |-- rootcause.go
|   |   |-- repair.go
|   |   |-- governance.go
|   |   \-- pragent.go
|   |
|   |-- services/             # Core integrations (LLM API & GitHub SDK)
|   |   |-- github.go
|   |   \-- llm.go
|   |
|   |-- models/               # Structs to maintain uniform payload structures
|   |   \-- event.go
|   |
|   |-- config/               # Environment Configuration loader
|   |   \-- config.go
|   |
|   \-- orchestrator/         # The coordination engine
|       \-- orchestrator.go
|
|-- prompts/                  # System prompts feeding instructions to the LLM
|   |-- root_cause_prompt.txt
|   |-- repair_prompt.txt
|   \-- governance_prompt.txt
|
|-- docker-compose.yml        # Docker compose containing App + Postgres
|-- Dockerfile                # Rootless secure multi-stage builder
|-- go.mod / go.sum           # Dependecy manager
\-- README.md                 # Project documentation
```

---

## Future Enhancements

- **Dashboard:** Create a React/Next.js frontend to visualize AI confidence scores and approve manual governed tasks.
- **Persistent State:** Log LLM outputs to PostgreSQL to retrain/fine-tune custom internal models.
- **Slack Integration:** Send direct actionable buttons via Slack for `HIGH` risk governance tasks.
- **Predictive CI/CD:** Predict flaky tests and pipeline failures *before* a workflow is even executed by running the Root Cause Analysis early in the git pre-commit hook.

---
*Built for Agentic DevOps engineering.*
