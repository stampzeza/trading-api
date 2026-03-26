package auth

import (
	"context"

	"trading-api/internal/db"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	c.BindJSON(&req)

	var id string
	var hash string

	err := db.DB.QueryRow(context.Background(),
		`SELECT id, password_hash FROM users WHERE email=$1`,
		req.Email,
	).Scan(&id, &hash)

	if err != nil {
		c.JSON(401, gin.H{"error": "invalid"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password))
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid"})
		return
	}

	token, _ := GenerateToken(id)

	c.JSON(200, gin.H{
		"token": token,
	})
}
