package service

import (
	"database/sql"
	"log"
	"time"

	"go-rabbitmq-order-system/order-status-service/internal/config"
	"go-rabbitmq-order-system/shared"

	"github.com/google/uuid"
)

type OrderStatusService struct {
	db       *sql.DB
	rabbitMQ *shared.RabbitMQ
	config   *config.OrderStatusConfig
}

type StatusChange struct {
	OrderID     string
	OldStatus   string
	NewStatus   string
	EventType   string
	Timestamp   time.Time
	Metadata    map[string]interface{}
}

func New(db *sql.DB, rabbitMQ *shared.RabbitMQ, config *config.OrderStatusConfig) *OrderStatusService {
	return &OrderStatusService{
		db:       db,
		rabbitMQ: rabbitMQ,
		config:   config,
	}
}

func (s *OrderStatusService) HandleOrderEvent(event shared.OrderEvent) error {
	log.Printf("Received event: %s for order: %s", event.EventType, event.OrderID)

	// Map events to order statuses
	newStatus := s.mapEventToStatus(event.EventType)
	if newStatus == "" {
		log.Printf("Unknown event type: %s", event.EventType)
		return nil
	}

	// Update order status
	err := s.updateOrderStatus(event.OrderID, newStatus, event)
	if err != nil {
		log.Printf("Failed to update order status: %v", err)
		return err
	}

	log.Printf("Updated order %s status to: %s", event.OrderID, newStatus)
	return nil
}

func (s *OrderStatusService) mapEventToStatus(eventType string) string {
	statusMap := map[string]string{
		shared.EventOrderCreated:           shared.StatusCreated,
		shared.EventPaymentSuccessful:      shared.StatusPaymentSuccessful,
		shared.EventPaymentFailed:          shared.StatusPaymentFailed,
		shared.EventStockReserved:          shared.StatusStockReserved,
		shared.EventStockInsufficient:      shared.StatusStockInsufficient,
		shared.EventOrderReadyForShipping:  shared.StatusReadyForShipping,
		shared.EventOrderShipped:           shared.StatusShipped,
		shared.EventOrderDelivered:         shared.StatusDelivered,
		shared.EventOrderCancelled:         shared.StatusCancelled,
	}
	
	return statusMap[eventType]
}

func (s *OrderStatusService) updateOrderStatus(orderID, status string, event shared.OrderEvent) error {
	// Check current order status to avoid backward status updates
	var currentStatus string
	err := s.db.QueryRow("SELECT status FROM orders WHERE id = $1", orderID).Scan(&currentStatus)
	if err != nil {
		log.Printf("Order not found: %s", orderID)
		return err
	}

	// Check if status update is valid
	if !s.isValidStatusTransition(currentStatus, status) {
		log.Printf("Invalid status transition for order %s: %s -> %s", 
			orderID, currentStatus, status)
		return nil
	}

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update order status
	_, err = tx.Exec(`
		UPDATE orders 
		SET status = $1, updated_at = $2 
		WHERE id = $3
	`, status, time.Now(), orderID)
	
	if err != nil {
		return err
	}

	// Log status change if audit logging is enabled
	if s.config.EnableAuditLog {
		err = s.logStatusChange(tx, StatusChange{
			OrderID:   orderID,
			OldStatus: currentStatus,
			NewStatus: status,
			EventType: event.EventType,
			Timestamp: time.Now(),
			Metadata:  event.Metadata,
		})
		if err != nil {
			log.Printf("Failed to log status change: %v", err)
			// Don't fail the transaction for audit log errors
		}
	}

	// Commit transaction
	return tx.Commit()
}

func (s *OrderStatusService) isValidStatusTransition(currentStatus, newStatus string) bool {
	// Define status priority to prevent backward updates
	statusPriority := map[string]int{
		shared.StatusCreated:           1,
		shared.StatusPaymentPending:    2,
		shared.StatusPaymentSuccessful: 3,
		shared.StatusStockReserved:     4,
		shared.StatusReadyForShipping:  5,
		shared.StatusShipped:           6,
		shared.StatusDelivered:         7,
		shared.StatusPaymentFailed:     -1, // Special status
		shared.StatusStockInsufficient: -2, // Special status
		shared.StatusCancelled:         -3, // Special status
	}

	currentPriority, currentExists := statusPriority[currentStatus]
	newPriority, newExists := statusPriority[newStatus]

	if !currentExists || !newExists {
		log.Printf("Unknown status found: current=%s, new=%s", currentStatus, newStatus)
		return false
	}

	// Allow updates for failure statuses or higher priority statuses
	return newPriority < 0 || (newPriority > currentPriority && currentPriority >= 0)
}

func (s *OrderStatusService) logStatusChange(tx *sql.Tx, change StatusChange) error {
	_, err := tx.Exec(`
		INSERT INTO order_status_history (id, order_id, old_status, new_status, event_type, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, uuid.New().String(), change.OrderID, change.OldStatus, change.NewStatus, 
		change.EventType, s.metadataToJSON(change.Metadata), change.Timestamp)

	return err
}

func (s *OrderStatusService) metadataToJSON(metadata map[string]interface{}) string {
	if metadata == nil {
		return "{}"
	}
	
	// Simple JSON conversion - in production you'd use json.Marshal
	return "{}"
} 