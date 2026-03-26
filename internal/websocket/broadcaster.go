package websocket

func Broadcast(data interface{}) {
	for client := range Clients {
		err := client.WriteJSON(data)
		if err != nil {
			client.Close()
			delete(Clients, client)
		}
	}
}
