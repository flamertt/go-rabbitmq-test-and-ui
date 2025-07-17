package service

import (
	"database/sql"
	"log"
	"math/rand"
	"time"

	"go-rabbitmq-order-system/payment-processing-service/internal/config"
	"go-rabbitmq-order-system/shared"

	"github.com/google/uuid"
)

type PaymentService struct {
	db       *sql.DB
	rabbitMQ *shared.RabbitMQ
	config   *config.PaymentGatewayConfig
}

type PaymentResult struct {
	Success       bool   `json:"success"`
	TransactionID string `json:"transaction_id"`
	Method        string `json:"method"`
	Message       string `json:"message"`
}

func New(db *sql.DB, rabbitMQ *shared.RabbitMQ, config *config.PaymentGatewayConfig) *PaymentService {
	return &PaymentService{
		db:       db,
		rabbitMQ: rabbitMQ,
		config:   config,
	}
}

func (s *PaymentService) HandleOrderEvent(event shared.OrderEvent) error {
	log.Printf("Received event: %s for order: %s", event.EventType, event.OrderID)

	if event.EventType != shared.EventOrderCreated {
		return nil
	}

	return s.processPayment(event)
}

func (s *PaymentService) processPayment(event shared.OrderEvent) error {
	log.Printf("Processing payment for order: %s, amount: %.2f", event.OrderID, event.TotalAmount)

	// Simulate payment processing
	result := s.simulatePayment(event.TotalAmount)

	// Store payment transaction
	err := s.storePaymentTransaction(event.OrderID, event.TotalAmount, result)
	if err != nil {
		log.Printf("Failed to store payment transaction: %v", err)
		return err
	}

	// Publish payment result event
	var eventType string
	if result.Success {
		eventType = shared.EventPaymentSuccessful
	} else {
		eventType = shared.EventPaymentFailed
	}

	resultEvent := shared.OrderEvent{
		EventType:   eventType,
		OrderID:     event.OrderID,
		UserID:      event.UserID,
		TotalAmount: event.TotalAmount,
		Status:      eventType,
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"transaction_id": result.TransactionID,
			"payment_method": result.Method,
			"message":        result.Message,
		},
	}

	return s.rabbitMQ.PublishEvent(resultEvent)
}

func (s *PaymentService) simulatePayment(amount float64) PaymentResult {
	// Simulate processing time
	time.Sleep(time.Duration(s.config.ProcessingDelayMS) * time.Millisecond)

	// Random payment methods
	methods := []string{"credit_card", "debit_card", "bank_transfer", "digital_wallet"}
	method := methods[rand.Intn(len(methods))]

	// Generate transaction ID
	transactionID := "TXN_" + uuid.New().String()[:8]

	// Simulate payment success/failure based on config
	success := rand.Float64() < s.config.SuccessRate

	var message string
	if success {
		message = "Payment processed successfully"
	} else {
		failureReasons := []string{
			"Insufficient funds",
			"Card expired",
			"Payment declined by bank",
			"Network timeout",
			"Invalid payment details",
		}
		message = failureReasons[rand.Intn(len(failureReasons))]
	}

	log.Printf("Payment simulation: Amount=%.2f, Method=%s, Success=%t, Message=%s", amount, method, success, message)

	return PaymentResult{
		Success:       success,
		TransactionID: transactionID,
		Method:        method,
		Message:       message,
	}
}

func (s *PaymentService) storePaymentTransaction(orderID string, amount float64, result PaymentResult) error {
	status := "FAILED"
	if result.Success {
		status = "SUCCESS"
	}

	_, err := s.db.Exec(`
		INSERT INTO payment_transactions (id, order_id, amount, status, transaction_id, payment_method, message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, uuid.New().String(), orderID, amount, status, result.TransactionID, result.Method, result.Message, time.Now())

	return err
} 