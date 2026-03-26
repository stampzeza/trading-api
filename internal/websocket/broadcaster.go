package websocket

import (
	"trading-api/internal/signal"
)

func Broadcast(data interface{}) {
	for client := range Clients {

		data := signal.GetAllSignals()

		filtered := signal.FilterSignals(client.UserID, data)

		msg := map[string]interface{}{
			"type": "INIT",
			"data": filtered,
		}

		client.Conn.WriteJSON(msg)
	}
}
