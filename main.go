package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
)

var db *pgx.Conn

// =========================
// WebSocket
// =========================

var clients = make(map[*websocket.Conn]bool)
var clientsMu sync.Mutex

var upgrader = websocket.Upgrader{
	CheckOrigin:       func(r *http.Request) bool { return true },
	EnableCompression: true,
}

// =========================
// Message Structure
// =========================

type MessageType string

const (
	TypeInit   MessageType = "INIT"
	TypeUpdate MessageType = "UPDATE"
)

type WSMessage struct {
	Type MessageType `json:"type"`
	Data interface{} `json:"data"`
}

// =========================
// Models
// =========================

type Trade struct {
	Symbol    string  `json:"symbol"`
	Type      string  `json:"type"`
	Lot       float64 `json:"lot"`
	Profit    float64 `json:"profit"`
	OpenTime  string  `json:"open_time"`
	CloseTime string  `json:"close_time"`
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

// =========================
// MAIN
// =========================

func main() {
	var err error

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
		port = "8080"
	}

	// 🔥 Start listener
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

	log.Println("🚀 Server running on :" + port)
	r.Run(":" + port)
}

// =========================
// DB Query (Clean)
// =========================

func fetchAllSignals() ([]TradeSignal, error) {
	rows, err := db.Query(context.Background(), `
		SELECT id, symbol, type, price, tp, sl, status, isactive, created_at 
		FROM "tbTradeSignal"
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
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
		if err != nil {
			continue
		}

		t.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
		results = append(results, t)
	}

	return results, nil
}

// =========================
// Broadcast
// =========================

func broadcast(msg WSMessage) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range clients {
		err := client.WriteJSON(msg)
		if err != nil {
			client.Close()
			delete(clients, client)
		}
	}
}

// =========================
// WebSocket Handler
// =========================

func wsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	log.Println("✅ Client connected")

	// 🔥 ส่ง INIT (ครั้งเดียว)
	go func() {
		data, err := fetchAllSignals()
		if err != nil {
			return
		}

		conn.WriteJSON(WSMessage{
			Type: TypeInit,
			Data: data,
		})
	}()

	// keep alive
	go func() {
		for {
			time.Sleep(20 * time.Second)
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				clientsMu.Lock()
				delete(clients, conn)
				clientsMu.Unlock()
				conn.Close()
				return
			}
		}
	}()

	// read loop
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			clientsMu.Lock()
			delete(clients, conn)
			clientsMu.Unlock()
			conn.Close()
			break
		}
	}
}

// =========================
// LISTEN / NOTIFY (Realtime)
// =========================

func listenSignalChanges() {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println("Listen DB error:", err)
		return
	}

	_, err = conn.Exec(context.Background(), "LISTEN signal_channel")
	if err != nil {
		log.Println("LISTEN error:", err)
		return
	}

	log.Println("👂 Listening for DB changes...")

	for {
		notification, err := conn.WaitForNotification(context.Background())
		if err != nil {
			log.Println("Wait error:", err)
			continue
		}

		var t TradeSignal
		err = json.Unmarshal([]byte(notification.Payload), &t)
		if err != nil {
			log.Println("Parse error:", err)
			continue
		}

		// 🔥 ส่งเฉพาะ row ที่เปลี่ยน
		broadcast(WSMessage{
			Type: TypeUpdate,
			Data: t,
		})
	}
}

// =========================
// API
// =========================

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

	c.JSON(200, gin.H{"status": "saved"})
}

func getTradeSignals(c *gin.Context) {
	data, err := fetchAllSignals()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, data)
}

// =========================
// CORS
// =========================

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
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
