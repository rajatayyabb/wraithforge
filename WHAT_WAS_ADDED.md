# Wraithforge — What Was Added

This builds on the original scaffold (see `WHAT_WAS_NOT_BUILT.pdf`). The scope
boundary from that document is unchanged: the autonomous-offensive core — an
exploit-writing/"coder" agent, attack-path planning, attack-chain replay, and
target-directed scanning or exploitation — was **not** added, in any framing.

What was added is the **defensive triage** path the original doc pointed at,
plus the engineering hardening the scaffold needed. Everything here is passive:
it ingests output from tools **you** run yourself and restructures it. Nothing
in this change connects to, scans, or acts against a target.

## 1. Triage ingestion (`backend/pkg/triage`)

A stdlib-only package that parses security-tool output into a normalized,
de-duplicated set of findings:

- `nmap.go` — parses `nmap -oX` XML. One informational "exposure" finding per
  open port, with the service banner and any NSE script output as evidence. It
  makes no exploitability judgement.
- `scanner.go` — parses generic scanner JSON (a `{tool, findings|results|
  vulnerabilities: [...]}` object, or a bare array). Accepts common field
  aliases so most tools map without a custom adapter.
- `triage.go` — the `Finding`/`Severity` model, severity parsing, and
  `Normalize()` (fills severity + fingerprint, de-dupes keeping the most severe
  copy, sorts most-severe-first).
- `triage_test.go` — unit tests for all of the above (no external deps, so
  `go test ./backend/pkg/triage/...` runs offline).

## 2. Findings persistence + API

- Migration `00003_findings.sql` — a `findings` table with a severity check
  constraint and a `UNIQUE (flow_id, fingerprint)` so repeated runs are
  idempotent.
- `pkg/database/findings.go` — `InsertFinding` (ON CONFLICT DO NOTHING),
  `ListFindingsByFlow` (severity-ordered), `FindingSeveritySummary`.
- `GET /api/flows/:flowId/findings` added to the REST API.
- The Markdown report (`pkg/report`) and the `report` CLI now include a
  Triage Findings section with a per-severity summary.

## 3. `triage` CLI (`backend/cmd/triage`)

```bash
# Dry run — parse, normalize, print a summary:
go run ./backend/cmd/triage -nmap backend/examples/nmap-sample.xml \
                            -scanner backend/examples/scanner-sample.json

# Persist against a flow (reads DATABASE_URL):
go run ./backend/cmd/triage -flow <flow-id> -db \
                            -nmap scan.xml -scanner findings.json
```

`-nmap` and `-scanner` are repeatable. Sample inputs are in
`backend/examples/`.

## 4. Real per-token auth

- `pkg/database/tokens.go` — `CreateAPIToken` (stores only a SHA-256 hash and
  returns the plaintext once), `ValidateAPIToken` (hash lookup, `revoked_at IS
  NULL`), `RevokeAPIToken`.
- `cmd/token` — mints a token: `go run ./backend/cmd/token -label dev`.
- `pkg/server/auth.go` rewritten:
  - constant-time comparison for the `API_TOKEN` dev fallback (`crypto/subtle`);
  - database-backed tokens via the previously-unused `api_tokens` table;
  - accepts the token from the `Authorization` header, a `session` cookie, **or**
    a `?token=` query parameter — the last is required because `EventSource`
    (the SSE log stream) cannot set headers. The original middleware rejected the
    stream for exactly this reason; this fixes it.
  - fails closed when no auth source is configured.

## 5. Sandbox hardening (`backend/pkg/sandbox`)

The generic Docker job runner now starts containers with:

- **no network** (`NetworkMode: none`, `NetworkDisabled`) by default —
  overridable via `Limits.AllowNetwork`;
- **read-only root filesystem**, with a small `noexec,nosuid` tmpfs at `/tmp`
  for scratch space;
- **all capabilities dropped** (`CapDrop: ALL`) and `no-new-privileges`;
- a **PID limit** (fork-bomb protection) and a **non-root user** (`nobody`) by
  default.

These are standard CI-runner isolation controls. Defaults are in
`DefaultLimits()`.

## 6. Frontend

- `providers/api.ts` — `Finding` type + `fetchFindings`.
- `features/findings/FindingsPanel.tsx` — severity-badged, collapsible findings
  list with a per-severity header tally.
- `App.tsx` — findings panel added beneath the task tree. `npm run build`
  (including `tsc`) passes.

## Build status

- **Frontend:** `npm install` + `npm run build` verified passing in this
  session.
- **Backend:** the triage package is stdlib-only and its tests are written to
  run offline. The full module (gin/pgx/docker) could not be compiled here
  because the Go module proxy is blocked under this network policy — the same
  limitation the original scaffold noted. Run `go mod tidy && go build ./... &&
  go test ./...` in an environment with normal network access to confirm.
