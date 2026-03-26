package main

import (
	"os"
	"trading-api/internal/billing"
	"trading-api/internal/db"
	"trading-api/internal/handler"
	"trading-api/internal/middleware"
	"trading-api/internal/websocket"

	"github.com/gin-gonic/gin"
)

func main() {
	db.Init()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // ใช้ตอน local
	}
	go websocket.ListenSignalChanges()
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())
	r.GET("/ws", websocket.HandleWS)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
	r.POST("/trade", handler.CreateTrade)
	r.POST("/addSignal", handler.CreateTradeSignal)
	r.POST("/updateSignal", handler.UpdateTradeSignal)
	r.GET("/signals", handler.CreateTradeSignal)

	r.POST("/create-checkout", billing.CreateCheckout)
	r.POST("/webhook", billing.StripeWebhook)

	r.Run(":" + port) // 👈 ต้องใช้แบบนี้
}
