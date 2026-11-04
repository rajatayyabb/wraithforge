package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"wraithforge/backend/pkg/database"
	"wraithforge/backend/pkg/report"
)

func main() {
	flowID := flag.String("flow", "", "flow id to report on")
	out := flag.String("out", "", "output markdown file (default: stdout)")
	flag.Parse()

	if *flowID == "" {
		fmt.Println("usage: report -flow <flow-id> [-out report.md]")
		os.Exit(1)
	}

	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		connString = "postgres://postgres:postgres@localhost:5432/wraithforge?sslmode=disable"
	}

	ctx := context.Background()
	db, err := database.Connect(ctx, connString)
	if err != nil {
		fmt.Println("connect error:", err)
		os.Exit(1)
	}
	defer db.Close()

	flows, err := db.ListFlows(ctx)
	if err != nil {
		fmt.Println("list flows error:", err)
		os.Exit(1)
	}
	var flow *database.Flow
	for i := range flows {
		if flows[i].ID == *flowID {
			flow = &flows[i]
			break
		}
	}
	if flow == nil {
		fmt.Println("flow not found:", *flowID)
		os.Exit(1)
	}

	results, err := db.ListResultsByFlow(ctx, *flowID)
	if err != nil {
		fmt.Println("list results error:", err)
		os.Exit(1)
	}

	costs, err := db.CostSummaryByFlow(ctx, *flowID)
	if err != nil {
		fmt.Println("cost summary error:", err)
		os.Exit(1)
	}

	md := report.RenderMarkdown(report.FlowReportData{Flow: *flow, Results: results, Costs: costs})

	if *out == "" {
		fmt.Print(md)
		return
	}
	if err := os.WriteFile(*out, []byte(md), 0o644); err != nil {
		fmt.Println("write error:", err)
		os.Exit(1)
	}
	fmt.Println("wrote", *out)
}
