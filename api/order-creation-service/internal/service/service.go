package service

import (
	"context"
	"time"

	"go-rabbitmq-order-system/order-creation-service/internal/repository"
	"go-rabbitmq-order-system/shared"

	"github.com/google/uuid"
)

type OrderService interface {
	CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error)
	GetOrder(ctx context.Context, orderID string) (*shared.Order, error)
	GetOrders(ctx context.Context, userID string) ([]shared.Order, error)
	GetProducts(ctx context.Context) ([]shared.Product, error)
	GetProduct(ctx context.Context, productID string) (*shared.Product, error)
}

type orderService struct {
	repo     repository.OrderRepository
	rabbitMQ *shared.RabbitMQ
}

type CreateOrderRequest struct {
	UserID string                    `json:"user_id" binding:"required"`
	Items  []CreateOrderItemRequest `json:"items" binding:"required,min=1"`
}

type CreateOrderItemRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

type CreateOrderResponse struct {
	OrderID     string  `json:"order_id"`
	UserID      string  `json:"user_id"`
	TotalAmount float64 `json:"total_amount"`
	Status      string  `json:"status"`
	Message     string  `json:"message"`
}

func New(repo repository.OrderRepository, rabbitMQ *shared.RabbitMQ) OrderService {
	return &orderService{
		repo:     repo,
		rabbitMQ: rabbitMQ,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	// Generate order ID
	orderID := uuid.New().String()

	// Calculate total amount and validate products
	var totalAmount float64
	var orderItems []shared.OrderItem

	for _, item := range req.Items {
		product, err := s.repo.GetProduct(ctx, item.ProductID)
		if err != nil {
			return nil, err
		}

		if product.StockQuantity < item.Quantity {
			return nil, ErrInsufficientStock
		}

		itemTotal := product.Price * float64(item.Quantity)
		totalAmount += itemTotal

		orderItems = append(orderItems, shared.OrderItem{
			ID:        uuid.New().String(),
			OrderID:   orderID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     product.Price,
		})
	}

	// Create order
	order := &shared.Order{
		ID:          orderID,
		UserID:      req.UserID,
		TotalAmount: totalAmount,
		Status:      shared.StatusCreated,
		Items:       orderItems,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateOrder(ctx, order); err != nil {
		return nil, err
	}

	// Publish order created event
	event := shared.OrderEvent{
		EventType:   shared.EventOrderCreated,
		OrderID:     orderID,
		UserID:      req.UserID,
		TotalAmount: totalAmount,
		Items:       orderItems,
		Status:      shared.StatusCreated,
		Timestamp:   time.Now(),
	}

	if err := s.rabbitMQ.PublishEvent(event); err != nil {
		// Log but don't fail the request
		// In production, you might want to use a retry mechanism
	}

	return &CreateOrderResponse{
		OrderID:     orderID,
		UserID:      req.UserID,
		TotalAmount: totalAmount,
		Status:      shared.StatusCreated,
		Message:     "Order created successfully",
	}, nil
}

func (s *orderService) GetOrder(ctx context.Context, orderID string) (*shared.Order, error) {
	return s.repo.GetOrder(ctx, orderID)
}

func (s *orderService) GetOrders(ctx context.Context, userID string) ([]shared.Order, error) {
	return s.repo.GetOrders(ctx, userID)
}

func (s *orderService) GetProducts(ctx context.Context) ([]shared.Product, error) {
	return s.repo.GetProducts(ctx)
}

func (s *orderService) GetProduct(ctx context.Context, productID string) (*shared.Product, error) {
	return s.repo.GetProduct(ctx, productID)
} 