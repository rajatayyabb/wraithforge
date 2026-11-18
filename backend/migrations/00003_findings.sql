-- +goose Up
CREATE TABLE findings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flow_id UUID NOT NULL REFERENCES flows(id) ON DELETE CASCADE,
    source_tool TEXT NOT NULL,
    asset TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL,
    severity TEXT NOT NULL DEFAULT 'info'
        CHECK (severity IN ('info', 'low', 'medium', 'high', 'critical')),
    category TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    evidence TEXT NOT NULL DEFAULT '',
    fingerprint TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (flow_id, fingerprint)
);

CREATE INDEX findings_flow_idx ON findings(flow_id);
CREATE INDEX findings_severity_idx ON findings(flow_id, severity);

-- +goose Down
DROP TABLE IF EXISTS findings;
