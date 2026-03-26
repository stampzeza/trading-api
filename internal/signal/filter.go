package signal

import (
	"context"
	"trading-api/internal/db"
)

func FilterSignals(userID *string, signals []TradeSignal) []TradeSignal {

	// ❌ ยังไม่ login
	if userID == nil {
		return onlyInactive(signals)
	}

	// 🔥 TODO: เช็ค subscription จริง (ตอนนี้ mock ไปก่อน)
	isSubscribed := checkSubscription(userID)

	if !isSubscribed {
		return onlyInactive(signals)
	}

	return signals
}

func onlyInactive(signals []TradeSignal) []TradeSignal {
	var result []TradeSignal

	for _, s := range signals {
		if !s.IsActive {
			result = append(result, s)
		}
	}

	return result
}

func checkSubscription(userID *string) bool {

	var status string

	err := db.DB.QueryRow(context.Background(),
		`SELECT status FROM subscriptions WHERE user_id=$1`,
		*userID,
	).Scan(&status)

	if err != nil {
		return false
	}

	return status == "active"
}
