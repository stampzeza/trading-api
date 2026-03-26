package billing

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"trading-api/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/webhook"
)

func StripeWebhook(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)

	event, err := webhook.ConstructEvent(
		body,
		c.GetHeader("Stripe-Signature"),
		"whsec_xxx",
	)

	if err != nil {
		c.JSON(400, gin.H{"error": "invalid webhook"})
		return
	}

	// 🔥 subscription created
	if event.Type == "checkout.session.completed" {

		var session stripe.CheckoutSession
		json.Unmarshal(event.Data.Raw, &session)

		userID := session.Metadata["user_id"]

		// save DB
		db.DB.Exec(context.Background(), `
			INSERT INTO subscriptions (user_id, status, stripe_customer_id)
			VALUES ($1,'active',$2)
		`, userID, session.Customer.ID)
	}

	// 🔥 subscription canceled
	if event.Type == "customer.subscription.deleted" {

		var sub stripe.Subscription
		json.Unmarshal(event.Data.Raw, &sub)

		db.DB.Exec(context.Background(), `
			UPDATE subscriptions SET status='inactive'
			WHERE stripe_subscription_id=$1
		`, sub.ID)
	}

	c.Status(http.StatusOK)
}
