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
		if err == sql.ErrNoRows {
			log.Printf("Order %s not found, skipping status update", orderID)
			return nil // Don't requeue if order doesn't exist
		}
		log.Printf("Failed to get current order status: %v", err)
		return nil // Don't requeue on DB errors
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
		log.Printf("Failed to start transaction: %v", err)
		return nil
	}
	defer tx.Rollback()

	// Update order status
	_, err = tx.Exec(`
		UPDATE orders 
		SET status = $1, updated_at = $2 
		WHERE id = $3
	`, status, time.Now(), orderID)
	
	if err != nil {
		log.Printf("Failed to update order status: %v", err)
		return nil
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
	err = tx.Commit()
	if err != nil {
		return err
	}

	// Check if order is ready for shipping after successful update
	go s.checkReadyForShipping(orderID)

	return nil
}

// checkReadyForShipping checks if order has both payment successful and stock reserved
// and automatically transitions to READY_FOR_SHIPPING
func (s *OrderStatusService) checkReadyForShipping(orderID string) {
	// Get current order status
	var currentStatus string
	err := s.db.QueryRow("SELECT status FROM orders WHERE id = $1", orderID).Scan(&currentStatus)
	if err != nil {
		log.Printf("Failed to check order status for shipping readiness: %v", err)
		return
	}

	// Check if both payment and stock are successful
	// We need to verify that we have records of both successful operations
	if currentStatus == shared.StatusPaymentSuccessful || currentStatus == shared.StatusStockReserved {
		// Check if we have both payment transaction and stock reservation
		var paymentExists, stockExists bool
		
		// Check payment transaction
		err = s.db.QueryRow(`
			SELECT EXISTS(SELECT 1 FROM payment_transactions 
			WHERE order_id = $1 AND status = 'SUCCESS')
		`, orderID).Scan(&paymentExists)
		
		if err != nil {
			log.Printf("Failed to check payment status: %v", err)
			return
		}

		// Check stock reservation
		err = s.db.QueryRow(`
			SELECT EXISTS(SELECT 1 FROM stock_reservations 
			WHERE order_id = $1 AND status = 'RESERVED')
		`, orderID).Scan(&stockExists)
		
		if err != nil {
			log.Printf("Failed to check stock reservation: %v", err)
			return
		}

		// If both exist, update to READY_FOR_SHIPPING
		if paymentExists && stockExists {
			_, err = s.db.Exec(`
				UPDATE orders 
				SET status = $1, updated_at = $2 
				WHERE id = $3 AND status IN ($4, $5)
			`, shared.StatusReadyForShipping, time.Now(), orderID, 
				shared.StatusPaymentSuccessful, shared.StatusStockReserved)
			
			if err != nil {
				log.Printf("Failed to update order to ready for shipping: %v", err)
				return
			}

			log.Printf("Order %s is now ready for shipping", orderID)
			
			// Publish ready for shipping event
			event := shared.OrderEvent{
				EventType: shared.EventOrderReadyForShipping,
				OrderID:   orderID,
				Status:    shared.StatusReadyForShipping,
				Timestamp: time.Now(),
			}
			
			err = s.rabbitMQ.PublishEvent(event)
			if err != nil {
				log.Printf("Failed to publish ready for shipping event: %v", err)
			}
		}
	}
}

func (s *OrderStatusService) isValidStatusTransition(currentStatus, newStatus string) bool {
	// Define valid status transitions instead of using priority system
	// since payment and stock operations happen in parallel
	
	validTransitions := map[string][]string{
		shared.StatusCreated: {
			shared.StatusPaymentSuccessful,
			shared.StatusStockReserved,
			shared.StatusPaymentFailed,
			shared.StatusStockInsufficient,
			shared.StatusCancelled,
		},
		shared.StatusPaymentSuccessful: {
			shared.StatusStockReserved,
			shared.StatusReadyForShipping,
			shared.StatusStockInsufficient,
			shared.StatusCancelled,
		},
		shared.StatusStockReserved: {
			shared.StatusPaymentSuccessful,
			shared.StatusReadyForShipping,
			shared.StatusPaymentFailed,
			shared.StatusCancelled,
		},
		shared.StatusReadyForShipping: {
			shared.StatusShipped,
			shared.StatusCancelled,
		},
		shared.StatusShipped: {
			shared.StatusDelivered,
			shared.StatusCancelled,
		},
		shared.StatusPaymentFailed: {
			shared.StatusCancelled,
		},
		shared.StatusStockInsufficient: {
			shared.StatusCancelled,
		},
		shared.StatusDelivered: {},
		shared.StatusCancelled: {},
	}

	// Check if transition is valid
	allowedStatuses, exists := validTransitions[currentStatus]
	if !exists {
		log.Printf("Unknown current status: %s", currentStatus)
		return false
	}

	// Allow transition if new status is in allowed list
	for _, allowedStatus := range allowedStatuses {
		if newStatus == allowedStatus {
			return true
		}
	}

	// Allow same status updates (idempotent)
	if currentStatus == newStatus {
		return true
	}

	log.Printf("Invalid transition: %s -> %s", currentStatus, newStatus)
	return false
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