package websocket

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin:       func(r *http.Request) bool { return true },
	EnableCompression: true,
}

func HandleWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	Clients[conn] = true
	log.Println("Client connected")

	go keepAlive(conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			conn.Close()
			delete(Clients, conn)
			break
		}
	}
}

func keepAlive(conn *websocket.Conn) {
	for {
		time.Sleep(20 * time.Second)

		err := conn.WriteMessage(websocket.PingMessage, nil)
		if err != nil {
			conn.Close()
			delete(Clients, conn)
			return
		}
	}
}
