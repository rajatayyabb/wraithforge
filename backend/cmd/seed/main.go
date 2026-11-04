package main

import (
	"context"
	"fmt"
	"os"

	"wraithforge/backend/pkg/database"
)

func main() {
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

	var flowIDs []string
	flows := []struct{ title, goal, status string }{
		{"Sample Flow A", "Demonstrate end-to-end orchestration", "completed"},
		{"Sample Flow B", "Second sample flow for pagination checks", "pending"},
		{"Sample Flow C", "Third sample flow", "running"},
	}
	for _, f := range flows {
		var id string
		err := db.Pool.QueryRow(ctx,
			`INSERT INTO flows (title, goal, status) VALUES ($1, $2, $3) RETURNING id`,
			f.title, f.goal, f.status,
		).Scan(&id)
		if err != nil {
			fmt.Println("insert flow error:", err)
			os.Exit(1)
		}
		flowIDs = append(flowIDs, id)
	}

	for i, flowID := range flowIDs {
		var taskID string
		err := db.Pool.QueryRow(ctx,
			`INSERT INTO tasks (flow_id, title, status, position) VALUES ($1, $2, $3, $4) RETURNING id`,
			flowID, fmt.Sprintf("Task %d.1", i+1), "pending", 0,
		).Scan(&taskID)
		if err != nil {
			fmt.Println("insert task error:", err)
			os.Exit(1)
		}

		_, err = db.Pool.Exec(ctx,
			`INSERT INTO subtasks (task_id, title, status, position) VALUES ($1, $2, $3, $4)`,
			taskID, fmt.Sprintf("Subtask %d.1.1", i+1), "pending", 0,
		)
		if err != nil {
			fmt.Println("insert subtask error:", err)
			os.Exit(1)
		}

		_, err = db.InsertAgentLog(ctx, flowID, "planner", "initialized flow")
		if err != nil {
			fmt.Println("insert log error:", err)
			os.Exit(1)
		}
	}

	fmt.Println("seed complete")
}
