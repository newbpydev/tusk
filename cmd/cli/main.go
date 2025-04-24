package main

import (
	"fmt"
	"log"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/lib/pq"
	"github.com/newbpydev/tusk/internal/adapters/db"
	"github.com/newbpydev/tusk/internal/config"
)

func main() {
	cfg := config.Load()
	fmt.Println("DB_URL:", cfg.DBURL)
	fmt.Println("PORT:", cfg.Port)
	fmt.Println("APP_ENV:", cfg.AppEnv)

	// Initialize the database connection
	// This will use the DSN from the environment variable DB_URL.
	if err := db.Connect(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()
}
