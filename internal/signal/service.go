package signal

import (
	"context"
	"time"

	"trading-api/internal/db"
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

func GetAllSignals() []TradeSignal {
	rows, err := db.DB.Query(context.Background(), `
		SELECT id, symbol, type, price, tp, sl, status, isactive, created_at 
		FROM "tbTradeSignal"
		ORDER BY id DESC
	`)
	if err != nil {
		return nil
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

	return results
}
