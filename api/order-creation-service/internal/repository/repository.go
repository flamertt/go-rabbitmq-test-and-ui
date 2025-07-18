package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"go-rabbitmq-order-system/shared"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrOrderNotFound   = errors.New("order not found")
)

type ProductsFilter struct {
	Search   string
	Category string
	MinPrice *float64
	MaxPrice *float64
}

type PaginationParams struct {
	Page     int
	PageSize int
	SortBy   string
	SortDir  string
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	Total      int64       `json:"total"`
	TotalPages int         `json:"total_pages"`
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *shared.Order) error
	GetOrder(ctx context.Context, orderID string) (*shared.Order, error)
	GetOrders(ctx context.Context, userID string) ([]shared.Order, error)
	GetProducts(ctx context.Context, filter *ProductsFilter, pagination *PaginationParams) (*PaginatedResponse, error)
	GetProduct(ctx context.Context, productID string) (*shared.Product, error)
}

type orderRepository struct {
	db *sql.DB
}

func New(db *sql.DB) OrderRepository {
	return &orderRepository{
		db: db,
	}
}

func (r *orderRepository) CreateOrder(ctx context.Context, order *shared.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert order
	_, err = tx.ExecContext(ctx,
		"INSERT INTO orders (id, user_id, total_amount, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		order.ID, order.UserID, order.TotalAmount, order.Status, order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		return err
	}

	// Insert order items
	for _, item := range order.Items {
		_, err = tx.ExecContext(ctx,
			"INSERT INTO order_items (id, order_id, product_id, quantity, price, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
			item.ID, item.OrderID, item.ProductID, item.Quantity, item.Price, order.CreatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *orderRepository) GetOrder(ctx context.Context, orderID string) (*shared.Order, error) {
	var order shared.Order
	err := r.db.QueryRowContext(ctx,
		"SELECT id, user_id, total_amount, status, created_at, updated_at FROM orders WHERE id = $1",
		orderID,
	).Scan(&order.ID, &order.UserID, &order.TotalAmount, &order.Status, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	// Get order items
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, order_id, product_id, quantity, price FROM order_items WHERE order_id = $1",
		orderID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []shared.OrderItem
	for rows.Next() {
		var item shared.OrderItem
		err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	order.Items = items
	return &order, nil
}

func (r *orderRepository) GetOrders(ctx context.Context, userID string) ([]shared.Order, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, user_id, total_amount, status, created_at, updated_at FROM orders WHERE user_id = $1 ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []shared.Order
	for rows.Next() {
		var order shared.Order
		err := rows.Scan(&order.ID, &order.UserID, &order.TotalAmount, &order.Status, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *orderRepository) GetProducts(ctx context.Context, filter *ProductsFilter, pagination *PaginationParams) (*PaginatedResponse, error) {
	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if filter.Search != "" {
		whereClause += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex+1)
		searchTerm := "%" + filter.Search + "%"
		args = append(args, searchTerm, searchTerm)
		argIndex += 2
	}

	if filter.Category != "" {
		whereClause += fmt.Sprintf(" AND category = $%d", argIndex)
		args = append(args, filter.Category)
		argIndex++
	}

	if filter.MinPrice != nil {
		whereClause += fmt.Sprintf(" AND price >= $%d", argIndex)
		args = append(args, *filter.MinPrice)
		argIndex++
	}

	if filter.MaxPrice != nil {
		whereClause += fmt.Sprintf(" AND price <= $%d", argIndex)
		args = append(args, *filter.MaxPrice)
		argIndex++
	}

	// Count total records
	countQuery := "SELECT COUNT(*) FROM products " + whereClause
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Build ORDER BY clause
	orderClause := "ORDER BY "
	switch pagination.SortBy {
	case "name":
		orderClause += "name"
	case "price":
		orderClause += "price"
	case "stock":
		orderClause += "stock_quantity"
	case "created_at":
		orderClause += "created_at"
	default:
		orderClause += "created_at"
	}

	if pagination.SortDir == "desc" {
		orderClause += " DESC"
	} else {
		orderClause += " ASC"
	}

	// Calculate offset
	offset := (pagination.Page - 1) * pagination.PageSize

	// Build final query
	query := fmt.Sprintf(`
		SELECT id, name, description, price, stock_quantity 
		FROM products %s %s 
		LIMIT $%d OFFSET $%d
	`, whereClause, orderClause, argIndex, argIndex+1)
	
	args = append(args, pagination.PageSize, offset)

	// Execute query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []shared.Product
	for rows.Next() {
		var product shared.Product
		err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.StockQuantity)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Calculate total pages
	totalPages := int((total + int64(pagination.PageSize) - 1) / int64(pagination.PageSize))

	return &PaginatedResponse{
		Data:       products,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (r *orderRepository) GetProduct(ctx context.Context, productID string) (*shared.Product, error) {
	var product shared.Product
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, description, price, stock_quantity FROM products WHERE id = $1",
		productID,
	).Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.StockQuantity)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	return &product, nil
} 