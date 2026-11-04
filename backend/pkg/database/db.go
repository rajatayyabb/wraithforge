package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func Connect(ctx context.Context, connString string) (*DB, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}
	return &DB{Pool: pool}, nil
}

func (db *DB) Close() {
	db.Pool.Close()
}

type Flow struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Goal      string    `json:"goal"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Task struct {
	ID       string `json:"id"`
	FlowID   string `json:"flowId"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Position int    `json:"position"`
}

type Subtask struct {
	ID       string `json:"id"`
	TaskID   string `json:"taskId"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Position int    `json:"position"`
}

type AgentLog struct {
	ID        string    `json:"id"`
	FlowID    string    `json:"flowId"`
	AgentRole string    `json:"agentRole"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

func (db *DB) ListFlows(ctx context.Context) ([]Flow, error) {
	rows, err := db.Pool.Query(ctx, `SELECT id, title, goal, status, created_at, updated_at FROM flows ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flows []Flow
	for rows.Next() {
		var f Flow
		if err := rows.Scan(&f.ID, &f.Title, &f.Goal, &f.Status, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		flows = append(flows, f)
	}
	return flows, rows.Err()
}

func (db *DB) ListTasksByFlow(ctx context.Context, flowID string) ([]Task, error) {
	rows, err := db.Pool.Query(ctx, `SELECT id, flow_id, title, status, position FROM tasks WHERE flow_id = $1 ORDER BY position`, flowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.FlowID, &t.Title, &t.Status, &t.Position); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (db *DB) ListSubtasksByTask(ctx context.Context, taskID string) ([]Subtask, error) {
	rows, err := db.Pool.Query(ctx, `SELECT id, task_id, title, status, position FROM subtasks WHERE task_id = $1 ORDER BY position`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subtasks []Subtask
	for rows.Next() {
		var s Subtask
		if err := rows.Scan(&s.ID, &s.TaskID, &s.Title, &s.Status, &s.Position); err != nil {
			return nil, err
		}
		subtasks = append(subtasks, s)
	}
	return subtasks, rows.Err()
}

func (db *DB) ListAgentLogs(ctx context.Context, flowID string, since time.Time) ([]AgentLog, error) {
	rows, err := db.Pool.Query(ctx, `SELECT id, flow_id, agent_role, message, created_at FROM agent_logs WHERE flow_id = $1 AND created_at > $2 ORDER BY created_at`, flowID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []AgentLog
	for rows.Next() {
		var l AgentLog
		if err := rows.Scan(&l.ID, &l.FlowID, &l.AgentRole, &l.Message, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func (db *DB) InsertAgentLog(ctx context.Context, flowID, role, message string) (string, error) {
	var id string
	err := db.Pool.QueryRow(ctx,
		`INSERT INTO agent_logs (flow_id, agent_role, message) VALUES ($1, $2, $3) RETURNING id`,
		flowID, role, message,
	).Scan(&id)
	return id, err
}

type Result struct {
	ID        string    `json:"id"`
	FlowID    string    `json:"flowId"`
	TaskID    *string   `json:"taskId,omitempty"`
	Title     string    `json:"title"`
	Summary   string    `json:"summary"`
	Detail    string    `json:"detail,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type CostEntry struct {
	ID           string    `json:"id"`
	FlowID       string    `json:"flowId"`
	Model        string    `json:"model"`
	InputTokens  int64     `json:"inputTokens"`
	OutputTokens int64     `json:"outputTokens"`
	CostUSD      float64   `json:"costUsd"`
	CreatedAt    time.Time `json:"createdAt"`
}

type CostSummary struct {
	Model        string  `json:"model"`
	InputTokens  int64   `json:"inputTokens"`
	OutputTokens int64   `json:"outputTokens"`
	CostUSD      float64 `json:"costUsd"`
}

func (db *DB) ListResultsByFlow(ctx context.Context, flowID string) ([]Result, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, flow_id, task_id, title, summary, COALESCE(detail, ''), created_at
		 FROM results WHERE flow_id = $1 ORDER BY created_at`, flowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Result
	for rows.Next() {
		var r Result
		if err := rows.Scan(&r.ID, &r.FlowID, &r.TaskID, &r.Title, &r.Summary, &r.Detail, &r.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func (db *DB) InsertResult(ctx context.Context, flowID string, taskID *string, title, summary, detail string) (string, error) {
	var id string
	err := db.Pool.QueryRow(ctx,
		`INSERT INTO results (flow_id, task_id, title, summary, detail) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		flowID, taskID, title, summary, detail,
	).Scan(&id)
	return id, err
}

func (db *DB) InsertCostEntry(ctx context.Context, flowID, model string, inputTokens, outputTokens int64, costUSD float64) (string, error) {
	var id string
	err := db.Pool.QueryRow(ctx,
		`INSERT INTO cost_tracking (flow_id, model, input_tokens, output_tokens, cost_usd) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		flowID, model, inputTokens, outputTokens, costUSD,
	).Scan(&id)
	return id, err
}

func (db *DB) CostSummaryByFlow(ctx context.Context, flowID string) ([]CostSummary, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT model, SUM(input_tokens), SUM(output_tokens), SUM(cost_usd)
		 FROM cost_tracking WHERE flow_id = $1 GROUP BY model ORDER BY model`, flowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summary []CostSummary
	for rows.Next() {
		var s CostSummary
		if err := rows.Scan(&s.Model, &s.InputTokens, &s.OutputTokens, &s.CostUSD); err != nil {
			return nil, err
		}
		summary = append(summary, s)
	}
	return summary, rows.Err()
}
