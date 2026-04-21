package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"

	"github.com/michael/language-arena/backend/internal/migration"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: migrate <up|status>")
		os.Exit(1)
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://lingouser:lingopass@localhost:5432/lingodb?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to db: %v\n", err)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "up":
		if err := migration.Run(db); err != nil {
			fmt.Fprintf(os.Stderr, "migration failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("migrations applied successfully")

	case "status":
		statuses, err := migration.Status(db)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get status: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%-40s  %-8s  %s\n", "MIGRATION", "STATUS", "APPLIED AT")
		fmt.Printf("%-40s  %-8s  %s\n", "---", "---", "---")
		for _, s := range statuses {
			status := "pending"
			appliedAt := ""
			if s.Applied {
				status = "applied"
				if s.AppliedAt != nil {
					appliedAt = s.AppliedAt.Format("2006-01-02 15:04:05")
				}
			}
			fmt.Printf("%-40s  %-8s  %s\n", s.Name, status, appliedAt)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown command %q. use: up, status\n", os.Args[1])
		os.Exit(1)
	}
}
