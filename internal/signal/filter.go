package signal

func FilterSignals(userID *string, signals []TradeSignal) []TradeSignal {

	// ❌ ยังไม่ login
	if userID == nil {
		return onlyInactive(signals)
	}

	// 🔥 TODO: เช็ค subscription จริง (ตอนนี้ mock ไปก่อน)
	isSubscribed := false

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
