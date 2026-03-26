package websocket

import (
	"log"
	"net/http"
	"time"

	"trading-api/internal/auth"
	"trading-api/internal/signal"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin:       func(r *http.Request) bool { return true },
	EnableCompression: true,
}

func HandleWS(c *gin.Context) {
	tokenStr := c.Query("token")

	var userID *string

	if tokenStr != "" {
		claims, err := auth.VerifyToken(tokenStr)
		if err == nil {
			userID = &claims.UserID
		}
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{
		Conn:   conn,
		UserID: userID,
	}

	Clients[client] = true
	log.Println("Client connected")

	go keepAlive(client)

	go func() {
		data := signal.GetAllSignals()

		filtered := signal.FilterSignals(client.UserID, data)

		conn.WriteJSON(map[string]interface{}{
			"type": "INIT",
			"data": filtered,
		})
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			conn.Close()
			delete(Clients, client)
			break
		}
	}
}

func keepAlive(client *Client) {
	for {
		time.Sleep(20 * time.Second)

		err := client.Conn.WriteMessage(websocket.PingMessage, nil)
		if err != nil {
			client.Conn.Close()
			delete(Clients, client)
			return
		}
	}
}
