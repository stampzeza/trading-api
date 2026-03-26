package websocket

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func ListenSignalChanges() {
	conn, _ := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	conn.Exec(context.Background(), "LISTEN signal_channel")

	for {
		notification, _ := conn.WaitForNotification(context.Background())

		var payload interface{}
		json.Unmarshal([]byte(notification.Payload), &payload)

		msg := WSMessage{
			Type: "UPDATE",
			Data: payload,
		}

		Broadcast(msg)
		log.Println("Broadcasting to clients:", len(Clients))
	}
}
