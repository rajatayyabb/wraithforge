package report

import (
	"fmt"
	"strings"

	"wraithforge/backend/pkg/database"
)

type FlowReportData struct {
	Flow       database.Flow
	Results    []database.Result
	Findings   []database.Finding
	Severities []database.SeverityCount
	Costs      []database.CostSummary
}

// RenderMarkdown builds a plain Markdown report from a flow's recorded
// results and cost-tracking entries. It contains no offensive-security
// content by itself; what goes into `Results` is entirely up to the caller.
func RenderMarkdown(data FlowReportData) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s\n\n", data.Flow.Title)
	fmt.Fprintf(&b, "**Goal:** %s\n\n", data.Flow.Goal)
	fmt.Fprintf(&b, "**Status:** %s\n\n", data.Flow.Status)
	fmt.Fprintf(&b, "**Generated:** %s\n\n", data.Flow.UpdatedAt.Format("2006-01-02 15:04 MST"))

	b.WriteString("## Results\n\n")
	if len(data.Results) == 0 {
		b.WriteString("_No results recorded._\n\n")
	}
	for _, r := range data.Results {
		fmt.Fprintf(&b, "### %s\n\n", r.Title)
		fmt.Fprintf(&b, "%s\n\n", r.Summary)
		if r.Detail != "" {
			fmt.Fprintf(&b, "```\n%s\n```\n\n", r.Detail)
		}
	}

	b.WriteString("## Triage Findings\n\n")
	if len(data.Findings) == 0 {
		b.WriteString("_No findings recorded._\n\n")
	} else {
		if len(data.Severities) > 0 {
			b.WriteString("| Severity | Count |\n|---|---|\n")
			for _, s := range data.Severities {
				fmt.Fprintf(&b, "| %s | %d |\n", s.Severity, s.Count)
			}
			b.WriteString("\n")
		}
		for _, f := range data.Findings {
			fmt.Fprintf(&b, "### [%s] %s\n\n", strings.ToUpper(f.Severity), f.Title)
			if f.Asset != "" {
				fmt.Fprintf(&b, "**Asset:** %s  \n", f.Asset)
			}
			fmt.Fprintf(&b, "**Source:** %s", f.SourceTool)
			if f.Category != "" {
				fmt.Fprintf(&b, " · **Category:** %s", f.Category)
			}
			b.WriteString("\n\n")
			if f.Description != "" {
				fmt.Fprintf(&b, "%s\n\n", f.Description)
			}
			if f.Evidence != "" {
				fmt.Fprintf(&b, "```\n%s\n```\n\n", f.Evidence)
			}
		}
	}

	b.WriteString("## Cost Summary\n\n")
	if len(data.Costs) == 0 {
		b.WriteString("_No cost entries recorded._\n\n")
	} else {
		b.WriteString("| Model | Input Tokens | Output Tokens | Cost (USD) |\n")
		b.WriteString("|---|---|---|---|\n")
		var totalCost float64
		for _, c := range data.Costs {
			fmt.Fprintf(&b, "| %s | %d | %d | $%.4f |\n", c.Model, c.InputTokens, c.OutputTokens, c.CostUSD)
			totalCost += c.CostUSD
		}
		fmt.Fprintf(&b, "\n**Total: $%.4f**\n\n", totalCost)
	}

	return b.String()
}
