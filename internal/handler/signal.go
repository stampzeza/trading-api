package handler

import (
	"context"
	"time"

	"trading-api/internal/db"

	"github.com/gin-gonic/gin"
)

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

func CreateTradeSignal(c *gin.Context) {
	var t TradeSignal
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	_, err := db.DB.Exec(context.Background(),
		`INSERT INTO "tbTradeSignal" (symbol, type, price, tp, sl, isactive, status, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		t.Symbol, t.Type, t.Price, t.Tp, t.Sl, t.IsActive, t.Status, t.CreatedAt,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "saved"})
}

func GetTradeSignals(c *gin.Context) {
	rows, err := db.DB.Query(context.Background(), `
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

		rows.Scan(&t.ID, &t.Symbol, &t.Type, &t.Price, &t.Tp, &t.Sl, &t.Status, &t.IsActive, &createdAt)
		t.CreatedAt = createdAt.Format("2006-01-02 15:04:05")

		results = append(results, t)
	}

	c.JSON(200, results)
}

func UpdateTradeSignal(c *gin.Context) {
	var t TradeSignal
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	_, err := db.DB.Exec(context.Background(),
		`UPDATE "tbTradeSignal" SET status=$1, isactive=$2 WHERE tp=$3 and sl=$4`,
		t.Status, t.IsActive, t.Tp, t.Sl,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "saved Update Signal Success."})
}
