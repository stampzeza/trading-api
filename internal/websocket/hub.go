package websocket

import "github.com/gorilla/websocket"

type Client struct {
	Conn   *websocket.Conn
	UserID *string
}

var Clients = make(map[*Client]bool)
