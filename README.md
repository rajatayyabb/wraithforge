<div align="center">

# WraithForge

### Observe. Orchestrate. Explain.

A transparent control plane for multi-agent workflows: plan work, run bounded jobs, stream execution logs, track costs, and produce auditable reports.

[![Status: Scaffold](https://img.shields.io/badge/status-scaffold-a855f7?style=for-the-badge)](https://github.com/rajatayyabb/wraithforge)
[![Go](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![React](https://img.shields.io/badge/React-18-61DAFB?style=for-the-badge&logo=react&logoColor=111827)](https://react.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=for-the-badge&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Vite](https://img.shields.io/badge/Vite-5-646CFF?style=for-the-badge&logo=vite&logoColor=white)](https://vitejs.dev/)

</div>

---

## The idea

Agent systems are powerful, but power without visibility is difficult to operate. WraithForge is built around a simple principle: every workflow should be understandable while it is running and explainable after it finishes.

The current repository is the orchestration foundation: a Go API, PostgreSQL-backed workflow model, generic Docker job runner, live SSE logs, React dashboard, and Markdown reporting. It intentionally does **not** ship offensive-security logic or target-directed automation.

## What is here today

| Capability | Implementation |
|---|---|
| Workflow and task model | PostgreSQL migrations for flows, tasks, subtasks, logs, results, tokens, and costs |
| API and authentication | Gin REST API with bearer/cookie authentication |
| Live observability | Server-Sent Events log stream and task tree dashboard |
| Bounded execution | Generic Docker runner with timeout and resource limits |
| Reporting | Markdown reports from results and cost data |
| Extensibility | GraphQL SDL foundation and YAML taxonomy loader |

## Architecture

```mermaid
flowchart LR
    UI[React + Vite dashboard] -->|REST / SSE| API[Go Gin API]
    API --> DB[(PostgreSQL + pgvector)]
    API --> RUN[Sandbox runner]
    RUN --> JOB[Bounded Docker jobs]
    API --> REPORT[Markdown reports]
    API --> TAX[Asset taxonomy]
```

## The execution gate

A future WraithForge execution gate is designed to separate **what ran** from **what was proven**. A result can be recorded without being promoted to a verified conclusion until its expected evidence matches.

```mermaid
stateDiagram-v2
  [*] --> Ran
  Ran --> Failed: exit != 0
  Ran --> Matched: exit 0
  Matched --> Verified: expected evidence matches
  Matched --> Unverified: no evidence match
  Failed --> Unverified
  Verified --> [*]
  Unverified --> [*]
```

The current scaffold stores execution results and logs; verified-finding promotion is part of the planned findings layer below.

## Roadmap

The attached WraithForge direction expands the foundation toward a controlled, evidence-first security workflow. These components are **planned and not yet present in this repository**:

| Planned layer | Intended role |
|---|---|
| Agent authoring | LLM-assisted script generation behind compilation and review gates |
| Graph planning | Validated DAGs, topological execution, and playbook reuse |
| Target-aware sandbox specs | Explicit mounts, exit-code capture, network policy, and cleanup guarantees |
| Findings and verification | ATT&CK mapping, CVSS scoring, evidence matching, and gate reasons |
| Operations tooling | Doctor command, end-to-end runner, dashboard findings pane, and exports |

This separation is deliberate: the generic runner should remain understandable and safe before target-directed capabilities are introduced.

## Project layout

```text
wraithforge/
├── backend/
│   ├── cmd/
│   │   ├── orchestrator/    # API server
│   │   ├── seed/            # Sample-data seeder
│   │   ├── report/          # Flow report CLI
│   │   ├── sandboxtest/     # Sandbox smoke test
│   │   └── taxoncheck/      # Taxonomy validator
│   ├── migrations/          # PostgreSQL schema migrations
│   ├── pkg/                 # API, database, sandbox, reports, taxonomy
│   └── taxonomy/            # Taxonomy definitions
├── frontend/
│   ├── src/App.tsx          # Dashboard shell
│   ├── src/features/        # Flows and tasks UI
│   └── src/providers/       # API client
└── README.md
```

## Quick start

### Prerequisites

- Go **1.24+**
- Node.js **22+** and npm
- PostgreSQL with the `pgvector` extension
- Docker, for sandbox execution
- [`goose`](https://github.com/pressly/goose), for migration convenience

### Start PostgreSQL

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

### Start the API

```bash
export API_TOKEN=dev-token
go run ./backend/cmd/orchestrator
```

The API listens on `http://localhost:8080`.

### Start the dashboard

```bash
cd frontend
npm ci
VITE_API_TOKEN=dev-token npm run dev
```

Open the Vite URL shown in your terminal.

## Commands

| Command | Purpose |
|---|---|
| `npm run build` | Build the frontend for production |
| `go run ./backend/cmd/orchestrator` | Start the REST + SSE API |
| `go run ./backend/cmd/seed` | Insert sample workflow data |
| `go run ./backend/cmd/sandboxtest -cmd "echo hello"` | Exercise the sandbox runner |
| `go run ./backend/cmd/report -flow <flow-id> -out report.md` | Render a Markdown report |
| `go run ./backend/cmd/taxoncheck` | Validate the taxonomy file |

## Configuration

| Variable | Default | Purpose |
|---|---|---|
| `DATABASE_URL` | — | PostgreSQL connection string |
| `API_TOKEN` | — | Development bearer/cookie secret |
| `VITE_API_TOKEN` | — | Frontend token passed to the API client |

Before deploying beyond local development, replace the development token flow, review Docker isolation settings, lock down CORS and database access, and add production-grade secret management and audit controls.

## API surface

```text
GET /api/flows
GET /api/flows/:flow-id/tasks
GET /api/flows/:flow-id/agent-logs/stream?token=<token>
```

The repository also includes a GraphQL SDL foundation for flows, tasks, subtasks, results, and costs.

## Security and scope

WraithForge is a generic orchestration scaffold. It does not include exploit automation, credential harvesting, target-specific attack workflows, or offensive-security logic. Use the sandbox only for controlled jobs in environments you own or are authorized to test.

Any future target-directed features should preserve the project's core guarantees: explicit scope, bounded execution, evidence-backed conclusions, auditable gate reasons, and reliable cleanup on success, error, or timeout.

## Project status

The frontend builds cleanly with `npm run build`. Backend dependency resolution and compilation should be verified in an environment with normal Go module network access. Production hardening, broader tests, richer workflow controls, and deployment automation remain open areas for development.

## Contributing

Focused pull requests and thoughtful issues are welcome. Keep changes small, document behavior changes, and preserve the emphasis on visibility, containment, and auditability.

## License

WraithForge is released under the [MIT License](LICENSE).
