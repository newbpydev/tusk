package main

import (
	"context"
	"fmt"
	"log"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5/pgtype"
	_ "github.com/lib/pq"

	dbAdapter "github.com/newbpydev/tusk/internal/adapters/db"
	sqlc "github.com/newbpydev/tusk/internal/adapters/db/sqlc"
	"github.com/newbpydev/tusk/internal/config"
)

func main() {
	ctx := context.Background()

	// Load the configuration
	// This will load the configuration from the environment variables and the config file.
	cfg := config.Load()
	fmt.Println("Config loaded: ", cfg)

	// Initialize the database connection
	// This will use the DSN from the environment variable DB_URL.
	if err := dbAdapter.Connect(ctx); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbAdapter.Close()

	// Get a queries instance
	q := sqlc.New(dbAdapter.Pool)

	// Create a dummy user
	user, err := q.CreateUser(ctx, sqlc.CreateUserParams{
		Username:     "testuser",
		Email:        "testuser@example.com",
		PasswordHash: "hashedpassword",
	})
	if err != nil {
		log.Fatalf("failed to create user: %v", err)
	}
	fmt.Printf("User created: %+v\n", user)

	// create a test task (no parent)
	taskParams := sqlc.CreateTaskParams{
		UserID:   user.ID,
		ParentID: pgtype.Int4{Int32: 0, Valid: false}, // root task
		Title:    "Test Task",
		Description: pgtype.Text{
			String: "Verify DB+sqlc",
			Valid:  true,
		},
		DueDate: pgtype.Timestamp{
			Time:  time.Now().Add(24 * time.Hour),
			Valid: true,
		},
		Status: pgtype.Text{
			String: "todo",
			Valid:  true,
		},
		Priority: pgtype.Text{
			String: "medium",
			Valid:  true,
		},
		IsCompleted: pgtype.Bool{
			Bool:  false,
			Valid: true,
		},
		Tags:         []string{"test", "task"},
		DisplayOrder: pgtype.Int4{Int32: 1, Valid: true},
	}

	task, err := q.CreateTask(ctx, taskParams)
	if err != nil {
		log.Fatalf("failed to create task: %v", err)
	}
	fmt.Printf("Task created: %+v\n", task)

}
