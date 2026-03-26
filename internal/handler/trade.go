package handler

import (
	"context"

	"trading-api/internal/db"

	"github.com/gin-gonic/gin"
)

type Trade struct {
	Symbol    string  `json:"symbol"`
	Type      string  `json:"type"`
	Lot       float64 `json:"lot"`
	Profit    float64 `json:"profit"`
	OpenTime  string  `json:"open_time"`
	CloseTime string  `json:"close_time"`
}

func CreateTrade(c *gin.Context) {
	var t Trade

	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	_, err := db.DB.Exec(context.Background(),
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
