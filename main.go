package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
)

var db *pgx.Conn
var clients = make(map[*websocket.Conn]bool)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	EnableCompression: true, // 🔥 เพิ่มตัวนี้
}

func main() {
	var err error

	// 👉 ใส่ตรงนี้เลย
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("❌ DATABASE_URL not set")
	}

	db, err = pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatal("❌ DB connect error:", err)
	}

	log.Println("✅ Connected to DB")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // ใช้ตอน local
	}
	go listenSignalChanges()
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(CORSMiddleware())
	r.GET("/ws", wsHandler)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
	r.POST("/trade", createTrade)
	r.POST("/addSignal", createTradeSignal)
	r.GET("/signals", getTradeSignals)

	r.Run(":" + port) // 👈 ต้องใช้แบบนี้
}

type Trade struct {
	Symbol    string  `json:"symbol"`
	Type      string  `json:"type"`
	Lot       float64 `json:"lot"`
	Profit    float64 `json:"profit"`
	OpenTime  string  `json:"open_time"`
	CloseTime string  `json:"close_time"`
}

func createTrade(c *gin.Context) {
	var t Trade

	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec(context.Background(),
		`INSERT INTO trades (symbol, type, lot, profit, open_time, close_time)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		t.Symbol, t.Type, t.Lot, t.Profit, t.OpenTime, t.CloseTime,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "saved"})
}

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

func createTradeSignal(c *gin.Context) {
	var t TradeSignal
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec(context.Background(),
		`INSERT INTO "tbTradeSignal" (symbol, type, price, tp, sl, isactive, status, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		t.Symbol, t.Type, t.Price, t.Tp, t.Sl, t.IsActive, t.Status, t.CreatedAt,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "saved Create Signal Success."})
}

func getTradeSignals(c *gin.Context) {
	rows, err := db.Query(context.Background(), `
		SELECT id, symbol, type, price, tp, sl, status, isactive, created_at 
		FROM "tbTradeSignal"
		ORDER BY id DESC
	`)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []TradeSignal

	for rows.Next() {
		var t TradeSignal
		var createdAt time.Time

		err := rows.Scan(
			&t.ID,
			&t.Symbol,
			&t.Type,
			&t.Price,
			&t.Tp,
			&t.Sl,
			&t.Status,
			&t.IsActive,
			&createdAt,
		)
		t.CreatedAt = createdAt.Format("2006-01-02 15:04:05")

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		results = append(results, t)
	}

	c.JSON(200, results)
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		// 🔥 ข้าม WebSocket
		if c.Request.URL.Path == "/ws" {
			c.Next()
			return
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
func wsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	clients[conn] = true
	log.Println("✅ Client connected")

	// 🔥 ส่งข้อมูลครั้งแรก
	go func() {
		pushLatestSignalsToClient(conn)
	}()

	// keep alive
	go func() {
		for {
			time.Sleep(20 * time.Second)

			err := conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				conn.Close()
				delete(clients, conn)
				return
			}
		}
	}()

	// read loop
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			conn.Close()
			delete(clients, conn)
			break
		}
	}
}
func broadcastSignals() {
	for {
		rows, err := db.Query(context.Background(), `
			SELECT id, symbol, type, price, tp, sl, status, isactive, created_at 
			FROM "tbTradeSignal"
			ORDER BY id DESC
		`)
		if err != nil {
			log.Println("DB error:", err)
			time.Sleep(3 * time.Second)
			continue
		}

		var results []TradeSignal

		for rows.Next() {
			var t TradeSignal
			var createdAt time.Time

			rows.Scan(
				&t.ID,
				&t.Symbol,
				&t.Type,
				&t.Price,
				&t.Tp,
				&t.Sl,
				&t.Status,
				&t.IsActive,
				&createdAt,
			)

			t.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
			results = append(results, t)
		}

		rows.Close()

		// 🔥 ยิงให้ทุก client
		for client := range clients {
			err := client.WriteJSON(results)
			if err != nil {
				log.Println("Write error:", err)
				client.Close()
				delete(clients, client)
			}
		}

		time.Sleep(3 * time.Second) // 🔥 realtime ทุก 3 วิ
	}
}
func listenSignalChanges() {
	conn, _ := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	conn.Exec(context.Background(), "LISTEN signal_channel")

	for {
		notification, _ := conn.WaitForNotification(context.Background())

		var signal TradeSignal
		err := json.Unmarshal([]byte(notification.Payload), &signal)
		if err != nil {
			log.Println("Parse payload error:", err)
			continue
		}

		msg := WSMessage{
			Type: "UPDATE",
			Data: signal,
		}

		broadcast(msg)
	}
}
func broadcast(msg WSMessage) {
	for client := range clients {
		err := client.WriteJSON(msg)
		if err != nil {
			client.Close()
			delete(clients, client)
		}
	}
}
func pushLatestSignals() {
	rows, err := db.Query(context.Background(), `
		SELECT id, symbol, type, price, tp, sl, status, isactive, created_at 
		FROM "tbTradeSignal"
		ORDER BY id DESC
	`)
	if err != nil {
		log.Println("DB error:", err)
		return
	}
	defer rows.Close()

	var results []TradeSignal

	for rows.Next() {
		var t TradeSignal
		var createdAt time.Time

		rows.Scan(
			&t.ID,
			&t.Symbol,
			&t.Type,
			&t.Price,
			&t.Tp,
			&t.Sl,
			&t.Status,
			&t.IsActive,
			&createdAt,
		)

		t.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
		results = append(results, t)
	}

	// 🔥 ยิงหา client
	for client := range clients {
		err := client.WriteJSON(results)
		if err != nil {
			client.Close()
			delete(clients, client)
		}
	}
}
func pushLatestSignalsToClient(conn *websocket.Conn) {
	data := getAllSignals()

	msg := WSMessage{
		Type: "INIT",
		Data: data,
	}

	conn.WriteJSON(msg)
}
func getAllSignals() []TradeSignal {
	rows, _ := db.Query(context.Background(), `
		SELECT id, symbol, type, price, tp, sl, status, isactive, created_at 
		FROM "tbTradeSignal"
		ORDER BY id DESC
	`)
	defer rows.Close()

	var results []TradeSignal

	for rows.Next() {
		var t TradeSignal
		var createdAt time.Time

		rows.Scan(
			&t.ID,
			&t.Symbol,
			&t.Type,
			&t.Price,
			&t.Tp,
			&t.Sl,
			&t.Status,
			&t.IsActive,
			&createdAt,
		)

		t.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
		results = append(results, t)
	}

	return results
}

type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
