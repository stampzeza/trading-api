package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

var db *pgx.Conn

func main() {
	var err error

	// 👉 ใส่ตรงนี้เลย
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("❌ DATABASE_URL not set")
	}

	db, err = pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatal("❌ DB connect error:", err)
	}

	log.Println("✅ Connected to DB")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // ใช้ตอน local
	}

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
	r.POST("/trade", createTrade)
	r.POST("/addSignal", createTradeSignal)
	r.GET("/signals", getTradeSignals)

	r.Run(":" + port) // 👈 ต้องใช้แบบนี้
}

type Trade struct {
	Symbol    string  `json:"symbol"`
	Type      string  `json:"type"`
	Lot       float64 `json:"lot"`
	Profit    float64 `json:"profit"`
	OpenTime  string  `json:"open_time"`
	CloseTime string  `json:"close_time"`
}

func createTrade(c *gin.Context) {
	var t Trade

	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec(context.Background(),
		`INSERT INTO trades (symbol, type, lot, profit, open_time, close_time)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		t.Symbol, t.Type, t.Lot, t.Profit, t.OpenTime, t.CloseTime,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "saved"})
}

type TradeSignal struct {
	ID        int     `json:"id"`
	Symbol    string  `json:"symbol"`
	Type      string  `json:"type"`
	Price     float64 `json:"price"`
	Tp        float64 `json:"tp"`
	Sl        float64 `json:"sl"`
	IsActive  bool    `json:"isActive"`
	Status    int     `json:"status"`
	CreatedAt string  `json:"created_at"`
}

func createTradeSignal(c *gin.Context) {
	var t TradeSignal
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec(context.Background(),
		`INSERT INTO "tbTradeSignal" (symbol, type, price, tp, sl, isactive, status, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		t.Symbol, t.Type, t.Price, t.Tp, t.Sl, t.IsActive, t.Status, t.CreatedAt,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "saved Create Signal Success."})
}

func getTradeSignals(c *gin.Context) {
	rows, err := db.Query(context.Background(), `
		SELECT id, symbol, type, price, tp, sl, status, isactive, created_at 
		FROM "tbTradeSignal"
		ORDER BY id DESC
	`)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []TradeSignal

	for rows.Next() {
		var t TradeSignal
		var createdAt time.Time

		err := rows.Scan(
			&t.ID,
			&t.Symbol,
			&t.Type,
			&t.Price,
			&t.Tp,
			&t.Sl,
			&t.Status,
			&t.IsActive,
			&createdAt,
		)
		t.CreatedAt = createdAt.Format("2006-01-02 15:04:05")

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		results = append(results, t)
	}

	c.JSON(200, results)
}
