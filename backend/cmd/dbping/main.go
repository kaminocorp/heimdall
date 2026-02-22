package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

func main() {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		fmt.Println("FAIL: DATABASE_URL is not set")
		os.Exit(1)
	}
	fmt.Println("Connecting...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		fmt.Printf("FAIL: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	var result int
	err = conn.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		fmt.Printf("FAIL: query error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("OK: connected successfully (SELECT 1 = %d)\n", result)
}
