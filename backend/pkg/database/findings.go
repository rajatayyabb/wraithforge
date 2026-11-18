package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// Finding is a normalized triage finding persisted against a flow. It mirrors
// the shape produced by pkg/triage, with database identity and timestamp.
type Finding struct {
	ID          string    `json:"id"`
	FlowID      string    `json:"flowId"`
	SourceTool  string    `json:"sourceTool"`
	Asset       string    `json:"asset"`
	Title       string    `json:"title"`
	Severity    string    `json:"severity"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Evidence    string    `json:"evidence"`
	Fingerprint string    `json:"fingerprint"`
	CreatedAt   time.Time `json:"createdAt"`
}

// SeverityCount is one row of the per-severity tally for a flow.
type SeverityCount struct {
	Severity string `json:"severity"`
	Count    int    `json:"count"`
}

// InsertFinding stores a finding, skipping it if an identical one (same
// flow_id + fingerprint) already exists. It returns inserted=false when the
// finding was a duplicate. This makes repeated triage runs idempotent.
func (db *DB) InsertFinding(ctx context.Context, flowID string, f Finding) (inserted bool, id string, err error) {
	err = db.Pool.QueryRow(ctx,
		`INSERT INTO findings
		    (flow_id, source_tool, asset, title, severity, category, description, evidence, fingerprint)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT (flow_id, fingerprint) DO NOTHING
		 RETURNING id`,
		flowID, f.SourceTool, f.Asset, f.Title, f.Severity, f.Category, f.Description, f.Evidence, f.Fingerprint,
	).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}
	return true, id, nil
}

// ListFindingsByFlow returns a flow's findings, most severe first.
func (db *DB) ListFindingsByFlow(ctx context.Context, flowID string) ([]Finding, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, flow_id, source_tool, asset, title, severity, category, description, evidence, fingerprint, created_at
		 FROM findings
		 WHERE flow_id = $1
		 ORDER BY
		    CASE severity
		        WHEN 'critical' THEN 4
		        WHEN 'high' THEN 3
		        WHEN 'medium' THEN 2
		        WHEN 'low' THEN 1
		        ELSE 0
		    END DESC,
		    asset, title`, flowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var findings []Finding
	for rows.Next() {
		var f Finding
		if err := rows.Scan(&f.ID, &f.FlowID, &f.SourceTool, &f.Asset, &f.Title,
			&f.Severity, &f.Category, &f.Description, &f.Evidence, &f.Fingerprint, &f.CreatedAt); err != nil {
			return nil, err
		}
		findings = append(findings, f)
	}
	return findings, rows.Err()
}

// FindingSeveritySummary returns per-severity counts for a flow.
func (db *DB) FindingSeveritySummary(ctx context.Context, flowID string) ([]SeverityCount, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT severity, COUNT(*)
		 FROM findings WHERE flow_id = $1
		 GROUP BY severity
		 ORDER BY
		    CASE severity
		        WHEN 'critical' THEN 4
		        WHEN 'high' THEN 3
		        WHEN 'medium' THEN 2
		        WHEN 'low' THEN 1
		        ELSE 0
		    END DESC`, flowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counts []SeverityCount
	for rows.Next() {
		var sc SeverityCount
		if err := rows.Scan(&sc.Severity, &sc.Count); err != nil {
			return nil, err
		}
		counts = append(counts, sc)
	}
	return counts, rows.Err()
}
