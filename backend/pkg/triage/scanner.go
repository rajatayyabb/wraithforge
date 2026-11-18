package triage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// scannerDoc is the flexible JSON shape ParseScannerJSON accepts. It covers the
// two most common layouts: a top-level object with a "findings" (or "results"/
// "vulnerabilities") array, or a bare array of finding objects. A "tool" field
// at the top level, if present, labels the source.
type scannerDoc struct {
	Tool            string         `json:"tool"`
	Findings        []scannerEntry `json:"findings"`
	Results         []scannerEntry `json:"results"`
	Vulnerabilities []scannerEntry `json:"vulnerabilities"`
}

// scannerEntry is one finding from a generic scanner. Several common field
// aliases are accepted so output from different tools maps without a custom
// adapter each time.
type scannerEntry struct {
	Target      string `json:"target"`
	Asset       string `json:"asset"`
	Package     string `json:"package"`
	Location    string `json:"location"`
	Title       string `json:"title"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Severity    string `json:"severity"`
	Category    string `json:"category"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Message     string `json:"message"`
	Evidence    string `json:"evidence"`
	Detail      string `json:"detail"`
}

func (e scannerEntry) asset() string {
	return firstNonEmpty(e.Asset, e.Target, e.Package, e.Location)
}

func (e scannerEntry) title() string {
	return firstNonEmpty(e.Title, e.ID, e.Name, "Unnamed finding")
}

func (e scannerEntry) category() string {
	return firstNonEmpty(e.Category, e.Type)
}

func (e scannerEntry) description() string {
	return firstNonEmpty(e.Description, e.Message)
}

func (e scannerEntry) evidence() string {
	return firstNonEmpty(e.Evidence, e.Detail)
}

// ParseScannerJSON reads generic scanner output (from a tool the operator ran)
// and returns normalized Findings. It accepts either a wrapper object with a
// findings/results/vulnerabilities array, or a bare top-level array. The
// severity strings are mapped via ParseSeverity. This is pure data
// restructuring — it performs no scanning of its own.
func ParseScannerJSON(data []byte) ([]Finding, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("parse scanner json: empty input")
	}

	var entries []scannerEntry
	tool := "scanner"

	switch trimmed[0] {
	case '[':
		if err := json.Unmarshal(trimmed, &entries); err != nil {
			return nil, fmt.Errorf("parse scanner json array: %w", err)
		}
	case '{':
		var doc scannerDoc
		if err := json.Unmarshal(trimmed, &doc); err != nil {
			return nil, fmt.Errorf("parse scanner json object: %w", err)
		}
		if doc.Tool != "" {
			tool = doc.Tool
		}
		entries = append(entries, doc.Findings...)
		entries = append(entries, doc.Results...)
		entries = append(entries, doc.Vulnerabilities...)
	default:
		return nil, fmt.Errorf("parse scanner json: expected object or array")
	}

	findings := make([]Finding, 0, len(entries))
	for _, e := range entries {
		findings = append(findings, Finding{
			SourceTool:  tool,
			Asset:       e.asset(),
			Title:       e.title(),
			Severity:    ParseSeverity(e.Severity),
			Category:    firstNonEmpty(e.category(), "finding"),
			Description: e.description(),
			Evidence:    e.evidence(),
		})
	}
	return findings, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
