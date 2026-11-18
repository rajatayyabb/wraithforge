// Command triage ingests security-tool output that the operator produced
// themselves (nmap -oX XML, generic scanner JSON) and turns it into a
// normalized, de-duplicated set of findings. With -db it also persists the
// findings against a flow so the dashboard and report pick them up.
//
// It reads files. It does not scan, connect to, or act against any target.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"wraithforge/backend/pkg/database"
	"wraithforge/backend/pkg/triage"
)

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

func main() {
	var nmapFiles multiFlag
	var scannerFiles multiFlag
	flag.Var(&nmapFiles, "nmap", "path to an nmap -oX XML file (repeatable)")
	flag.Var(&scannerFiles, "scanner", "path to a generic scanner JSON file (repeatable)")
	flowID := flag.String("flow", "", "flow id to attach findings to (required with -db)")
	persist := flag.Bool("db", false, "persist findings to the database for this flow")
	flag.Parse()

	if len(nmapFiles) == 0 && len(scannerFiles) == 0 {
		fmt.Println("usage: triage [-nmap scan.xml ...] [-scanner findings.json ...] [-flow <id> -db]")
		os.Exit(1)
	}

	var all []triage.Finding

	for _, path := range nmapFiles {
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("read %s: %v\n", path, err)
			os.Exit(1)
		}
		findings, err := triage.ParseNmapXML(data)
		if err != nil {
			fmt.Printf("parse %s: %v\n", path, err)
			os.Exit(1)
		}
		all = append(all, findings...)
	}

	for _, path := range scannerFiles {
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("read %s: %v\n", path, err)
			os.Exit(1)
		}
		findings, err := triage.ParseScannerJSON(data)
		if err != nil {
			fmt.Printf("parse %s: %v\n", path, err)
			os.Exit(1)
		}
		all = append(all, findings...)
	}

	normalized := triage.Normalize(all)
	printSummary(normalized)

	if !*persist {
		return
	}

	if *flowID == "" {
		fmt.Println("error: -db requires -flow <id>")
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

	var inserted, skipped int
	for _, f := range normalized {
		ok, _, err := db.InsertFinding(ctx, *flowID, database.Finding{
			SourceTool:  f.SourceTool,
			Asset:       f.Asset,
			Title:       f.Title,
			Severity:    string(f.Severity),
			Category:    f.Category,
			Description: f.Description,
			Evidence:    f.Evidence,
			Fingerprint: f.Fingerprint,
		})
		if err != nil {
			fmt.Println("insert finding error:", err)
			os.Exit(1)
		}
		if ok {
			inserted++
		} else {
			skipped++
		}
	}
	fmt.Printf("\npersisted to flow %s: %d new, %d already present\n", *flowID, inserted, skipped)
}

func printSummary(findings []triage.Finding) {
	counts := triage.Counts(findings)
	fmt.Printf("%d finding(s) after normalization:\n", len(findings))
	for _, sev := range []triage.Severity{
		triage.SeverityCritical, triage.SeverityHigh, triage.SeverityMedium,
		triage.SeverityLow, triage.SeverityInfo,
	} {
		if counts[sev] > 0 {
			fmt.Printf("  %-9s %d\n", sev, counts[sev])
		}
	}
	fmt.Println()
	for _, f := range findings {
		fmt.Printf("[%-8s] %-28s %s\n", strings.ToUpper(string(f.Severity)), f.Asset, f.Title)
	}
}
