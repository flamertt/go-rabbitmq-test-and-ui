package service

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"go-rabbitmq-order-system/stock-reservation-service/internal/config"
	"go-rabbitmq-order-system/shared"

	"github.com/google/uuid"
)

type StockService struct {
	db       *sql.DB
	rabbitMQ *shared.RabbitMQ
	config   *config.StockReservationConfig
}

type StockReservationResult struct {
	Success      bool               `json:"success"`
	Message      string             `json:"message"`
	Reservations []StockReservation `json:"reservations,omitempty"`
}

type StockReservation struct {
	ProductID     string `json:"product_id"`
	Quantity      int    `json:"quantity"`
	ReservationID string `json:"reservation_id"`
}

func New(db *sql.DB, rabbitMQ *shared.RabbitMQ, config *config.StockReservationConfig) *StockService {
	return &StockService{
		db:       db,
		rabbitMQ: rabbitMQ,
		config:   config,
	}
}

func (s *StockService) HandleOrderEvent(event shared.OrderEvent) error {
	log.Printf("Received event: %s for order: %s", event.EventType, event.OrderID)

	// Only process order created events
	if event.EventType != shared.EventOrderCreated {
		return nil
	}

	return s.processStockReservation(event)
}

func (s *StockService) processStockReservation(event shared.OrderEvent) error {
	log.Printf("Processing stock reservation for order: %s", event.OrderID)

	// Reserve stock for order items
	result := s.reserveStock(event)

	// Publish stock reservation result event
	var eventType string
	if result.Success {
		eventType = shared.EventStockReserved
	} else {
		eventType = shared.EventStockInsufficient
	}

	resultEvent := shared.OrderEvent{
		EventType:   eventType,
		OrderID:     event.OrderID,
		UserID:      event.UserID,
		TotalAmount: event.TotalAmount,
		Items:       event.Items,
		Status:      eventType,
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"message":      result.Message,
			"reservations": result.Reservations,
		},
	}

	return s.rabbitMQ.PublishEvent(resultEvent)
}

func (s *StockService) reserveStock(event shared.OrderEvent) StockReservationResult {
	// Start transaction with retry logic
	var result StockReservationResult
	
	for attempt := 0; attempt < s.config.RetryAttempts; attempt++ {
		result = s.attemptStockReservation(event)
		if result.Success {
			break
		}
		
		if attempt < s.config.RetryAttempts-1 {
			log.Printf("Stock reservation attempt %d failed, retrying...", attempt+1)
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}
	
	return result
}

func (s *StockService) attemptStockReservation(event shared.OrderEvent) StockReservationResult {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		log.Printf("Failed to start transaction: %v", err)
		return StockReservationResult{
			Success: false,
			Message: "Failed to start transaction",
		}
	}
	defer tx.Rollback()

	var reservations []StockReservation
	var insufficientProducts []string

	// Check and reserve stock for each item
	for _, item := range event.Items {
		log.Printf("Checking stock for product %s, quantity %d", item.ProductID, item.Quantity)

		// Get current stock with row lock (pessimistic locking)
		var currentStock int
		err := tx.QueryRow(`
			SELECT stock_quantity 
			FROM products 
			WHERE id = $1 
			FOR UPDATE
		`, item.ProductID).Scan(&currentStock)
		
		if err != nil {
			log.Printf("Product not found: %s", item.ProductID)
			return StockReservationResult{
				Success: false,
				Message: "Product not found: " + item.ProductID,
			}
		}

		// Check if sufficient stock available
		if currentStock < item.Quantity {
			log.Printf("Insufficient stock for product %s: required %d, available %d", 
				item.ProductID, item.Quantity, currentStock)
			insufficientProducts = append(insufficientProducts, item.ProductID)
			continue
		}

		// Reserve stock by updating product quantity
		_, err = tx.Exec(`
			UPDATE products 
			SET stock_quantity = stock_quantity - $1, updated_at = $2 
			WHERE id = $3
		`, item.Quantity, time.Now(), item.ProductID)
		
		if err != nil {
			log.Printf("Failed to update stock for product %s: %v", item.ProductID, err)
			return StockReservationResult{
				Success: false,
				Message: "Failed to update stock for product: " + item.ProductID,
			}
		}

		// Create stock reservation record
		reservationID := uuid.New().String()
		_, err = tx.Exec(`
			INSERT INTO stock_reservations (id, order_id, product_id, quantity, status, created_at, expires_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, reservationID, event.OrderID, item.ProductID, item.Quantity, "RESERVED", 
			time.Now(), time.Now().Add(time.Duration(s.config.ReservationTimeoutMinutes)*time.Minute))
		
		if err != nil {
			log.Printf("Failed to create stock reservation: %v", err)
			return StockReservationResult{
				Success: false,
				Message: "Failed to create stock reservation",
			}
		}

		reservations = append(reservations, StockReservation{
			ProductID:     item.ProductID,
			Quantity:      item.Quantity,
			ReservationID: reservationID,
		})

		log.Printf("Stock reserved successfully for product %s: %d units", item.ProductID, item.Quantity)
	}

	// Check if there were any insufficient stock issues
	if len(insufficientProducts) > 0 {
		return StockReservationResult{
			Success: false,
			Message: "Insufficient stock for products: " + fmt.Sprintf("%v", insufficientProducts),
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return StockReservationResult{
			Success: false,
			Message: "Failed to commit stock reservation",
		}
	}

	log.Printf("Stock reservation completed successfully for order: %s", event.OrderID)
	return StockReservationResult{
		Success:      true,
		Message:      "Stock reserved successfully",
		Reservations: reservations,
	}
} 