# Agentic CI/CD Pipeline Repair & Intelligent Release Orchestration System

A production-grade "Self-Healing CI/CD Pipeline" built in Go. It monitors CI/CD pipelines, diagnoses failures using contextual signals (commit diffs, logs, test reports) via LLMs, generates fixes automatically, and enforces a human-in-the-loop governance layer before creating a Pull Request for the fix.

## Multi-Agent Architecture

1. **Pipeline Monitoring Agent:** Receives GitHub Actions webhooks and detects pipeline failures.
2. **Root Cause Analysis Agent:** Analyzes logs, commit diffs, and test reports with an LLM.
3. **Auto Repair Agent:** Generates a minimal fix/patch in unified diff format.
4. **Governance Agent:** Evaluates the risk (LOW, MEDIUM, HIGH) of the patch.
5. **Pull Request Agent:** Creates automated fix Pull Requests on GitHub.
6. **Orchestrator:** Coordinates the execution flow across agents.

## Tech Stack
- **Language:** Go (Golang)
- **Web Framework:** Gin
- **LLM Integration:** go-openai
- **GitHub Integration:** go-github (v69)
- **Logging:** Zap
- **Database:** PostgreSQL (via Docker)
- **Containerization:** Docker & Docker Compose

## API Endpoints
- `POST /webhook/github`: Receives GitHub Workflow Run Webhooks.
- `GET /health`: Health check endpoint.

## Getting Started

1. **Clone the repository:**
   ```bash
   git clone <repo-url>
   cd agentic-cicd
   ```

2. **Configuration:**
   Copy `.env.example` to `.env` and fill in your keys.
   ```bash
   cp .env.example .env
   ```
   You will need:
   - `GITHUB_TOKEN`: A Personal Access Token (classic) with `repo` and `workflow` permissions.
   - `LLM_API_KEY`: An OpenAI API Key.

3. **Running Locally (Docker Compose):**
   ```bash
   docker-compose up --build
   ```

4. **Testing Webhooks:**
   You can map the `/webhook/github` endpoint to the internet using `ngrok` to attach it to an actual GitHub Repository webhook.
   ```bash
   ngrok http 8080
   ```
   Then set up the Webhook in your Github repo: 
   Payload URL: `<ngrok_url>/webhook/github`
   Content-Type: `application/json`
   Events: `Workflow runs`

## Project Structure
`cmd/` - Contains the main server executable entry point.
`internal/agents/` - Houses the business logic for all 5 Agents.
`internal/services/` - Wrapper implementations mapping to GitHub & LLM SDKs.
`internal/orchestrator/` - The coordination engine tying the system together.
`internal/models/` - Structs to maintain uniform payload structures.
`prompts/` - The text prompts sent to the LLM agent APIs.
