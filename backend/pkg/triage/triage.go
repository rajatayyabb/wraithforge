// Package triage ingests output from security tools the operator runs
// themselves (e.g. nmap -oX, dependency/vulnerability scanners) and turns it
// into a normalized, de-duplicated set of Findings.
//
// It is deliberately passive: every parser in this package reads a file that
// already exists on disk. Nothing here connects to a target, scans a host, or
// executes tooling. It triages evidence the operator collected; it does not
// decide what to attack or carry anything out.
package triage

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

// Severity is a normalized severity label. Parsers map tool-specific strings
// onto these five levels via ParseSeverity.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// severityRank orders severities from least to most serious. Used for sorting
// and for keeping the most serious duplicate when two findings collide.
var severityRank = map[Severity]int{
	SeverityInfo:     0,
	SeverityLow:      1,
	SeverityMedium:   2,
	SeverityHigh:     3,
	SeverityCritical: 4,
}

// Rank returns the ordinal rank of a severity (info=0 .. critical=4).
func (s Severity) Rank() int {
	return severityRank[s]
}

// ParseSeverity maps a free-form severity string from a tool onto one of the
// canonical Severity values. Unknown or empty values become SeverityInfo so a
// finding is never silently dropped.
func ParseSeverity(raw string) Severity {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "critical", "crit":
		return SeverityCritical
	case "high", "important", "error":
		return SeverityHigh
	case "medium", "moderate", "warning", "warn":
		return SeverityMedium
	case "low", "minor":
		return SeverityLow
	case "info", "informational", "none", "unknown", "negligible", "":
		return SeverityInfo
	default:
		return SeverityInfo
	}
}

// Finding is a single normalized observation. It is intentionally generic so
// the same shape serves port/service exposure, a vulnerable dependency, a web
// scanner alert, and so on.
type Finding struct {
	SourceTool  string   `json:"sourceTool"`
	Asset       string   `json:"asset"`
	Title       string   `json:"title"`
	Severity    Severity `json:"severity"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Evidence    string   `json:"evidence"`
	// Fingerprint is a stable identity for the finding, used for
	// de-duplication. Normalize fills it in when it is empty.
	Fingerprint string `json:"fingerprint"`
}

// computeFingerprint derives a stable hash from the fields that define a
// finding's identity. Two findings with the same tool, asset, title and
// category are considered the same observation even if their free-text
// description or evidence differs.
func (f Finding) computeFingerprint() string {
	h := sha256.New()
	h.Write([]byte(strings.ToLower(f.SourceTool)))
	h.Write([]byte{0})
	h.Write([]byte(strings.ToLower(f.Asset)))
	h.Write([]byte{0})
	h.Write([]byte(strings.ToLower(f.Title)))
	h.Write([]byte{0})
	h.Write([]byte(strings.ToLower(f.Category)))
	return hex.EncodeToString(h.Sum(nil))
}

// SeverityCounts tallies findings by severity.
type SeverityCounts map[Severity]int

// Normalize cleans up a slice of findings: it fills in any missing severity
// and fingerprint, removes duplicates (keeping the most severe copy), and
// returns the result sorted most-severe first, then by asset and title.
func Normalize(in []Finding) []Finding {
	byFP := make(map[string]Finding, len(in))

	for _, f := range in {
		if f.Severity == "" {
			f.Severity = SeverityInfo
		}
		if f.Fingerprint == "" {
			f.Fingerprint = f.computeFingerprint()
		}
		existing, ok := byFP[f.Fingerprint]
		if !ok {
			byFP[f.Fingerprint] = f
			continue
		}
		// Duplicate: keep whichever is more severe; prefer the one with
		// richer evidence when severities are equal.
		if f.Severity.Rank() > existing.Severity.Rank() ||
			(f.Severity.Rank() == existing.Severity.Rank() && len(f.Evidence) > len(existing.Evidence)) {
			byFP[f.Fingerprint] = f
		}
	}

	out := make([]Finding, 0, len(byFP))
	for _, f := range byFP {
		out = append(out, f)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Severity.Rank() != out[j].Severity.Rank() {
			return out[i].Severity.Rank() > out[j].Severity.Rank()
		}
		if out[i].Asset != out[j].Asset {
			return out[i].Asset < out[j].Asset
		}
		return out[i].Title < out[j].Title
	})
	return out
}

// Counts tallies a slice of findings by severity.
func Counts(findings []Finding) SeverityCounts {
	c := SeverityCounts{}
	for _, f := range findings {
		c[f.Severity]++
	}
	return c
}
