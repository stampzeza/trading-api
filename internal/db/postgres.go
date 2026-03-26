package db

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

var DB *pgx.Conn

func Init() *pgx.Conn {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("❌ DATABASE_URL not set")
	}

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatal("❌ DB connect error:", err)
	}

	log.Println("✅ Connected to DB")
	DB = conn
	return conn
}
