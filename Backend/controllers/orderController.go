package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkoutsession"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Models
type Order struct {
	ID      uint    `gorm:"primaryKey" json:"id"`
	UserID  uint    `json:"user_id"`
	Items   string  `json:"items"` // Serialized items data (JSON string)
	Amount  float64 `json:"amount"`
	Address string  `json:"address"`
	Payment bool    `json:"payment"`
	Status  string  `json:"status"`
}

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	CartData string `json:"cart_data"` // Serialized cart data (JSON string)
}

var (
	db             *gorm.DB
	frontendURL    = "http://localhost:5174"
	stripeSecret   = os.Getenv("STRIPE_SECRET_KEY")
	stripeCurrency = "inr"
)

func init() {
	var err error
	// Initialize database
	db, err = gorm.Open(sqlite.Open("orders.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Migrate models
	db.AutoMigrate(&Order{}, &User{})

	// Set Stripe API key
	stripe.Key = stripeSecret
}

// PlaceOrder handles placing a new order and creating a Stripe session
func PlaceOrder(c *gin.Context) {
	var req struct {
		UserID  uint          `json:"user_id"`
		Items   []interface{} `json:"items"`
		Amount  float64       `json:"amount"`
		Address string        `json:"address"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input"})
		return
	}

	order := Order{
		UserID:  req.UserID,
		Items:   fmt.Sprintf("%v", req.Items), // Serialize items as a string
		Amount:  req.Amount,
		Address: req.Address,
	}

	// Save the order
	if err := db.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to create order"})
		return
	}

	// Update user's cart data (assuming this is correct)
	if err := db.Model(&User{}).Where("id = ?", req.UserID).Update("cart_data", "").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to update cart"})
		return
	}

	// Create Stripe session line items
	lineItems := []*stripe.CheckoutSessionLineItemParams{}
	for _, item := range req.Items {
		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String(stripeCurrency),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe.String("Item Name"), // Replace with actual item name
				},
				UnitAmount: stripe.Int64(10000), // Replace with actual item price
			},
			Quantity: stripe.Int64(1), // Replace with actual item quantity
		})
	}

	// Add delivery charges
	lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
		PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
			Currency: stripe.String(stripeCurrency),
			ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
				Name: stripe.String("Delivery Charges"),
			},
			UnitAmount: stripe.Int64(2000), // Replace with delivery charges
		},
		Quantity: stripe.Int64(1),
	})

	// Create Stripe checkout session
	sessionParams := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems:          lineItems,
		Mode:               stripe.String("payment"),
		SuccessURL:         stripe.String(fmt.Sprintf("%s/verify?success=true&orderId=%d", frontendURL, order.ID)),
		CancelURL:          stripe.String(fmt.Sprintf("%s/verify?success=false&orderId=%d", frontendURL, order.ID)),
	}
	session, err := checkoutsession.New(sessionParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to create Stripe session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "session_url": session.URL})
}

// VerifyOrder handles verifying order payment status
func VerifyOrder(c *gin.Context) {
	var req struct {
		OrderID uint `json:"order_id"`
		Success bool `json:"success"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input"})
		return
	}

	if req.Success {
		// Mark order as paid
		db.Model(&Order{}).Where("id = ?", req.OrderID).Update("payment", true)
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Order paid"})
	} else {
		// Delete the order
		db.Delete(&Order{}, req.OrderID)
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Order not paid"})
	}
}

// UserOrders retrieves orders for a specific user
func UserOrders(c *gin.Context) {
	var req struct {
		UserID uint `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input"})
		return
	}

	var orders []Order
	db.Where("user_id = ?", req.UserID).Find(&orders)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": orders})
}

// ListOrders lists all orders (for admin panel)
func ListOrders(c *gin.Context) {
	var orders []Order
	db.Find(&orders)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": orders})
}

// UpdateStatus updates the status of an order
func UpdateStatus(c *gin.Context) {
	var req struct {
		OrderID uint   `json:"order_id"`
		Status  string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input"})
		return
	}

	db.Model(&Order{}).Where("id = ?", req.OrderID).Update("status", req.Status)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Status updated"})
}

func main() {
	router := gin.Default()

	router.POST("/place-order", PlaceOrder)
	router.POST("/verify-order", VerifyOrder)
	router.POST("/user-orders", UserOrders)
	router.GET("/list-orders", ListOrders)
	router.POST("/update-status", UpdateStatus)

	router.Run(":8080")
}
