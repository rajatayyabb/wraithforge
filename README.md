<div align="center">

# WraithForge

### Observe. Orchestrate. Explain.

**A transparent operations cockpit for workflows, sandboxed jobs, and defensive security triage.**

[![Status: Beta Scaffold](https://img.shields.io/badge/status-beta%20scaffold-a855f7?style=for-the-badge)](https://github.com/rajatayyabb/wraithforge)
[![Go](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![React](https://img.shields.io/badge/React-18-61DAFB?style=for-the-badge&logo=react&logoColor=111827)](https://react.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=for-the-badge&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Vite](https://img.shields.io/badge/Vite-5-646CFF?style=for-the-badge&logo=vite&logoColor=white)](https://vitejs.dev/)

</div>

---

## The short version

WraithForge helps you answer three questions without digging through a pile of terminals:

1. **What is running?** Follow flows, tasks, subtasks, and live agent logs.
2. **What did the tools report?** Ingest output you generated yourself and normalize it into findings.
3. **What can a reviewer trust?** Keep evidence, severity, fingerprints, and report context together.

This release adds a **defensive triage path** on top of the original orchestration scaffold. It parses Nmap XML and scanner JSON that you provide, de-duplicates the results, persists them against a flow, and surfaces them in the API, dashboard, and Markdown report.

> WraithForge does not scan targets, exploit systems, or make exploitability judgments. It organizes and explains output from tools you run yourself in environments you own or are authorized to assess.

## What changed in v2

| Upgrade | Why it matters |
|---|---|
| **Triage ingestion** | Parse Nmap XML and flexible scanner JSON into one normalized finding model |
| **Finding fingerprints** | De-duplicate repeated observations while keeping the most severe copy |
| **Findings persistence** | Store findings per flow with idempotent database writes |
| **Findings API** | Fetch severity-ordered findings from `GET /api/flows/:flowId/findings` |
| **Findings dashboard** | Browse severity-badged, collapsible findings beneath the task tree |
| **Database-backed tokens** | Store only SHA-256 token hashes and return plaintext once |
| **SSE-friendly auth** | Support headers, cookies, and query tokens for `EventSource` |
| **Hardened sandbox defaults** | No network, read-only root filesystem, dropped capabilities, PID limits, and non-root execution |
| **Audit-ready reports** | Add a triage findings section with a per-severity summary |

## The v2 data path

```mermaid
flowchart LR
    N[Nmap XML] --> T[Triage normalizer]
    S[Scanner JSON] --> T
    T --> D[De-duplicate + fingerprint]
    D --> F[(Findings table)]
    F --> API[REST API]
    F --> R[Markdown report]
    API --> UI[Findings panel]
```

The system is intentionally passive: **ingest → normalize → persist → explain**.

## Core architecture

```mermaid
flowchart TB
    UI[React dashboard] -->|REST + SSE| API[Go Gin API]
    API --> DB[(PostgreSQL + pgvector)]
    API --> SB[Hardened Docker runner]
    SB --> JOB[Bounded generic job]
    API --> REP[Markdown reports]
    TRI[Defensive triage CLI] --> DB
    TRI --> REP
```

## Safety boundary

WraithForge is a defensive orchestration and triage scaffold. It does **not** contain:

- Exploit-writing or autonomous attack agents
- Attack-path planning or attack-chain replay
- Target-directed scanning or exploitation
- Credential harvesting or offensive payloads
- Automatic exploitability claims from open ports or scanner output

The triage layer only reads files you already produced. Nmap open ports become informational exposure findings; the system does not decide that an exposed service is exploitable.

## Quick start

### Prerequisites

- Go **1.24+**
- Node.js **22+** and npm
- PostgreSQL with the `pgvector` extension
- Docker for sandbox smoke tests
- [`goose`](https://github.com/pressly/goose) for convenient migrations

### 1. Start PostgreSQL

```bash
docker run --name wraithforge-db \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=wraithforge \
  -p 5432:5432 \
  -d pgvector/pgvector:pg16
```

```bash
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/wraithforge?sslmode=disable
goose -dir backend/migrations postgres "$DATABASE_URL" up
go run ./backend/cmd/seed
```

### 2. Create a development token

```bash
go run ./backend/cmd/token -label dev
```

The command prints the plaintext token once. For a quick local fallback, you can use:

```bash
export API_TOKEN=dev-token
```

### 3. Start the API and dashboard

```bash
go run ./backend/cmd/orchestrator
```

In another terminal:

```bash
cd frontend
npm ci
VITE_API_TOKEN=<token> npm run dev
```

Open the Vite URL shown in your terminal.

## Triage your own tool output

### Dry run: parse and summarize

Use the included examples to see the normalizer without touching the database:

```bash
go run ./backend/cmd/triage \
  -nmap backend/examples/nmap-sample.xml \
  -scanner backend/examples/scanner-sample.json
```

### Persist findings against a flow

```bash
go run ./backend/cmd/triage \
  -flow <flow-id> \
  -db \
  -nmap scan.xml \
  -scanner findings.json
```

Then render the report:

```bash
go run ./backend/cmd/report -flow <flow-id> -out report.md
```

Both `-nmap` and `-scanner` are repeatable. Scanner JSON accepts a common object shape such as `{ "tool": "...", "findings": [...] }`, as well as a bare findings array.

## Command center

| Command | Purpose |
|---|---|
| `go run ./backend/cmd/orchestrator` | Start the REST + SSE API |
| `go run ./backend/cmd/seed` | Insert sample workflow data |
| `go run ./backend/cmd/token -label dev` | Mint a database-backed API token |
| `go run ./backend/cmd/triage ...` | Parse, normalize, and optionally persist tool output |
| `go run ./backend/cmd/report -flow <id> -out report.md` | Render a flow report |
| `go run ./backend/cmd/sandboxtest -cmd "echo hello"` | Exercise the isolated sandbox |
| `go run ./backend/cmd/taxoncheck backend/taxonomy/taxonomy.yaml` | Validate the taxonomy |
| `cd frontend && npm run build` | Build the dashboard for production |

## Repository map

```text
wraithforge/
├── backend/
│   ├── cmd/
│   │   ├── orchestrator/    # REST + SSE API
│   │   ├── seed/            # Sample data
│   │   ├── token/           # API token minting
│   │   ├── triage/          # Tool-output ingestion CLI
│   │   ├── report/          # Markdown report CLI
│   │   ├── sandboxtest/     # Sandbox smoke test
│   │   └── taxoncheck/      # Taxonomy validator
│   ├── examples/            # Sample Nmap XML and scanner JSON
│   ├── migrations/          # PostgreSQL schema, including findings
│   └── pkg/
│       ├── database/        # Workflow, token, and findings persistence
│       ├── report/          # Report generation
│       ├── sandbox/         # Hardened generic Docker runner
│       ├── server/          # Authenticated API and SSE stream
│       └── triage/          # Stdlib-only parsers and normalizer
├── frontend/
│   └── src/features/findings/ # Severity-badged findings panel
├── WHAT_WAS_ADDED.md
└── README.md
```

## API surface

```text
GET /api/flows
GET /api/flows/:flowId/tasks
GET /api/flows/:flowId/findings
GET /api/flows/:flowId/agent-logs/stream?token=<token>
```

Findings are returned severity-first. The SSE endpoint accepts a query token because browser `EventSource` cannot attach an `Authorization` header.

## Sandbox defaults

The generic Docker runner defaults to controls suitable for bounded local jobs:

- Network disabled
- Read-only root filesystem
- Small `noexec,nosuid` temporary filesystem
- All Linux capabilities dropped
- `no-new-privileges`
- PID limit to reduce fork-bomb risk
- Non-root execution as `nobody`
- Cleanup on completion, error, and timeout

These controls are not a substitute for a dedicated production isolation boundary. Review them before running untrusted workloads.

## Configuration

| Variable | Default | Purpose |
|---|---|---|
| `DATABASE_URL` | — | PostgreSQL connection string |
| `API_TOKEN` | — | Development fallback secret |
| `VITE_API_TOKEN` | — | Frontend token passed to the API client |

Before deployment, replace the development fallback, restrict database access, configure CORS deliberately, and use a proper secret-management system.

## Verification status

- **Frontend:** `npm ci` and `npm run build` passed during this review, including TypeScript checking.
- **Go checks:** pending in this review environment because the Go compiler is unavailable.
- **Before publishing:** run `go test ./backend/pkg/triage/...` and then `go mod tidy && go build ./... && go test ./...` with Go 1.24+ installed.

## License

WraithForge is released under the [MIT License](LICENSE). If this ZIP is being used as a new release archive, copy the repository's `LICENSE` file into the project root before publishing.
