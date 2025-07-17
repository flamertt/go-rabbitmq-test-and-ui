package service

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"go-rabbitmq-order-system/shipping-service/internal/config"
	"go-rabbitmq-order-system/shared"

	"github.com/google/uuid"
)

type ShippingService struct {
	db       *sql.DB
	rabbitMQ *shared.RabbitMQ
	config   *config.ShippingConfig
}

type ShippingResult struct {
	Success        bool   `json:"success"`
	TrackingNumber string `json:"tracking_number"`
	Carrier        string `json:"carrier"`
	Message        string `json:"message"`
	EstimatedDays  int    `json:"estimated_days"`
}

func New(db *sql.DB, rabbitMQ *shared.RabbitMQ, config *config.ShippingConfig) *ShippingService {
	return &ShippingService{
		db:       db,
		rabbitMQ: rabbitMQ,
		config:   config,
	}
}

func (s *ShippingService) HandleOrderEvent(event shared.OrderEvent) error {
	log.Printf("Received event: %s for order: %s", event.EventType, event.OrderID)

	// Process different types of events
	switch event.EventType {
	case shared.EventPaymentSuccessful:
		return s.checkReadyForShipping(event.OrderID)
	case shared.EventStockReserved:
		return s.checkReadyForShipping(event.OrderID)
	default:
		// Ignore other events
		return nil
	}
}

func (s *ShippingService) checkReadyForShipping(orderID string) error {
	// Check if both payment is successful and stock is reserved
	var orderStatus string
	var totalAmount float64
	err := s.db.QueryRow(`
		SELECT status, total_amount 
		FROM orders 
		WHERE id = $1
	`, orderID).Scan(&orderStatus, &totalAmount)
	
	if err != nil {
		log.Printf("Failed to get order status: %v", err)
		return err
	}

	// Check payment status
	var paymentTransactionStatus string
	err = s.db.QueryRow(`
		SELECT status 
		FROM payment_transactions 
		WHERE order_id = $1
	`, orderID).Scan(&paymentTransactionStatus)
	
	if err != nil {
		log.Printf("Payment not found for order %s", orderID)
		return nil // Payment might not be processed yet
	}

	// Check stock reservation
	var stockReservationCount int
	err = s.db.QueryRow(`
		SELECT COUNT(*) 
		FROM stock_reservations 
		WHERE order_id = $1 AND status = 'RESERVED'
	`, orderID).Scan(&stockReservationCount)
	
	if err != nil {
		log.Printf("Failed to check stock reservations: %v", err)
		return err
	}

	// If both payment and stock reservation are successful, proceed with shipping
	if paymentTransactionStatus == "SUCCESS" && stockReservationCount > 0 {
		return s.processShipping(orderID, totalAmount)
	}

	log.Printf("Order %s not ready for shipping yet. Payment: %s, Stock reservations: %d", 
		orderID, paymentTransactionStatus, stockReservationCount)
	return nil
}

func (s *ShippingService) processShipping(orderID string, totalAmount float64) error {
	log.Printf("Processing shipping for order: %s, amount: %.2f", orderID, totalAmount)

	// Create shipment
	result := s.createShipment(totalAmount)

	// Store shipping information
	err := s.storeShippingInfo(orderID, result)
	if err != nil {
		log.Printf("Failed to store shipping info: %v", err)
		return err
	}

	// Publish shipping event
	shippingEvent := shared.OrderEvent{
		EventType:   shared.EventOrderShipped,
		OrderID:     orderID,
		TotalAmount: totalAmount,
		Status:      shared.EventOrderShipped,
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"tracking_number": result.TrackingNumber,
			"carrier":         result.Carrier,
			"estimated_days":  result.EstimatedDays,
			"message":         result.Message,
		},
	}

	return s.rabbitMQ.PublishEvent(shippingEvent)
}

func (s *ShippingService) createShipment(totalAmount float64) ShippingResult {
	// Simulate processing time
	time.Sleep(time.Duration(s.config.ProcessingDelayMinutes) * time.Minute)

	// Random carrier selection
	carrier := s.config.Carriers[rand.Intn(len(s.config.Carriers))]

	// Generate tracking number
	trackingNumber := fmt.Sprintf("%s_%s", carrier, uuid.New().String()[:8])

	// Estimate delivery days based on amount (premium shipping for higher amounts)
	var estimatedDays int
	if totalAmount > s.config.PremiumThreshold {
		estimatedDays = rand.Intn(2) + 1 // 1-2 days for premium
	} else if totalAmount > s.config.StandardThreshold {
		estimatedDays = rand.Intn(3) + 2 // 2-4 days for standard
	} else {
		estimatedDays = rand.Intn(5) + 3 // 3-7 days for economy
	}

	message := fmt.Sprintf("Package shipped via %s, estimated delivery in %d days", 
		carrier, estimatedDays)

	log.Printf("Shipping simulation: Amount=%.2f, Carrier=%s, Tracking=%s, Days=%d", 
		totalAmount, carrier, trackingNumber, estimatedDays)

	return ShippingResult{
		Success:        true,
		TrackingNumber: trackingNumber,
		Carrier:        carrier,
		Message:        message,
		EstimatedDays:  estimatedDays,
	}
}

func (s *ShippingService) storeShippingInfo(orderID string, result ShippingResult) error {
	_, err := s.db.Exec(`
		INSERT INTO shipping_info (id, order_id, tracking_number, carrier, estimated_delivery_days, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, uuid.New().String(), orderID, result.TrackingNumber, result.Carrier, 
		result.EstimatedDays, "SHIPPED", time.Now())

	return err
} 