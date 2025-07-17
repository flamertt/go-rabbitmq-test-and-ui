package repository

import (
	"context"
	"database/sql"
	"errors"

	"go-rabbitmq-order-system/shared"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrOrderNotFound   = errors.New("order not found")
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *shared.Order) error
	GetOrder(ctx context.Context, orderID string) (*shared.Order, error)
	GetProducts(ctx context.Context) ([]shared.Product, error)
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

func (r *orderRepository) GetProducts(ctx context.Context) ([]shared.Product, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, description, price, stock_quantity FROM products")
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

	return products, nil
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