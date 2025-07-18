package shared

import "time"

// Order represents the main order entity
type Order struct {
	ID          string    `json:"order_id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	TotalAmount float64   `json:"total_amount" db:"total_amount"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	Items       []OrderItem `json:"items"`
}

// OrderItem represents individual items in an order
type OrderItem struct {
	ID        string  `json:"id" db:"id"`
	OrderID   string  `json:"order_id" db:"order_id"`
	ProductID string  `json:"product_id" db:"product_id"`
	Quantity  int     `json:"quantity" db:"quantity"`
	Price     float64 `json:"price" db:"price"`
}

// OrderEvent represents events published to RabbitMQ
type OrderEvent struct {
	EventType   string    `json:"event_type"`
	OrderID     string    `json:"order_id"`
	UserID      string    `json:"user_id"`
	TotalAmount float64   `json:"total_amount"`
	Items       []OrderItem `json:"items,omitempty"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Product represents a product in the system
type Product struct {
	ID            string  `json:"id" db:"id"`
	Name          string  `json:"name" db:"name"`
	Description   string  `json:"description" db:"description"`
	Price         float64 `json:"price" db:"price"`
	StockQuantity int     `json:"stock_quantity" db:"stock_quantity"`
}

// Order statuses
const (
	StatusCreated           = "CREATED"
	StatusPaymentPending    = "PAYMENT_PENDING"
	StatusPaymentSuccessful = "PAYMENT_SUCCESSFUL"
	StatusPaymentFailed     = "PAYMENT_FAILED"
	StatusStockReserved     = "STOCK_RESERVED"
	StatusStockInsufficient = "STOCK_INSUFFICIENT"
	StatusReadyForShipping  = "READY_FOR_SHIPPING"
	StatusShipped           = "SHIPPED"
	StatusDelivered         = "DELIVERED"
	StatusCancelled         = "CANCELLED"
)

// Event types
const (
	EventOrderCreated           = "OrderCreated"
	EventPaymentSuccessful      = "PaymentSuccessful"
	EventPaymentFailed          = "PaymentFailed"
	EventStockReserved          = "StockReserved"
	EventStockInsufficient      = "StockInsufficient"
	EventOrderReadyForShipping  = "OrderReadyForShipping"
	EventOrderShipped           = "OrderShipped"
	EventOrderDelivered         = "OrderDelivered"
	EventOrderCancelled         = "OrderCancelled"
) 