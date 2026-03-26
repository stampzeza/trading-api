package signal

import (
	"context"
	"time"

	"trading-api/internal/db"
)

type TradeSignal struct {
	ID        int
	Symbol    string
	Type      string
	Price     float64
	Tp        float64
	Sl        float64
	IsActive  bool
	Status    int
	CreatedAt string
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
