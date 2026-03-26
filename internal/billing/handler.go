package billing

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/checkout/session"
)

func CreateCheckout(c *gin.Context) {
	stripe.Key = "sk_test_xxx"

	userID := c.GetString("userID") // จาก middleware

	params := &stripe.CheckoutSessionParams{
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String("http://localhost:3000/success"),
		CancelURL:  stripe.String("http://localhost:3000/cancel"),

		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String("price_xxx"),
				Quantity: stripe.Int64(1),
			},
		},

		Metadata: map[string]string{
			"user_id": userID, // 🔥 สำคัญมาก
		},
	}

	s, err := session.New(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(200, gin.H{"url": s.URL})
}
