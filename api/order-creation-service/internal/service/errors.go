package service

import "errors"

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrInsufficientStock  = errors.New("insufficient stock")
	ErrInvalidOrderStatus = errors.New("invalid order status")
) 