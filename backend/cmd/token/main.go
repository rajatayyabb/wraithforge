// Command token mints a database-backed API token. The plaintext token is
// printed once; only its hash is stored, so there is no way to recover it
// later. Use the printed value as the Bearer token / VITE_API_TOKEN.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"wraithforge/backend/pkg/database"
)

func main() {
	label := flag.String("label", "", "human-readable label for the token (required)")
	flag.Parse()

	if *label == "" {
		fmt.Println("usage: token -label <name>")
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

	token, id, err := db.CreateAPIToken(ctx, *label)
	if err != nil {
		fmt.Println("create token error:", err)
		os.Exit(1)
	}

	fmt.Printf("created token %s (label %q)\n", id, *label)
	fmt.Println("token (shown once, store it now):")
	fmt.Println(token)
}
