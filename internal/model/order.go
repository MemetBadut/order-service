package model

import (
    "time"

    "gorm.io/gorm"
)

type OrderStatus string

const (
    StatusPending   OrderStatus = "pending"
    StatusConfirmed OrderStatus = "confirmed"
    StatusCancelled OrderStatus = "cancelled"
)

type Order struct {
    ID         uint           `gorm:"primaryKey;autoIncrement" json:"id"`
    ProductID  uint           `gorm:"not null" json:"product_id"`
    Quantity   int            `gorm:"not null" json:"quantity"`
    TotalPrice float64        `gorm:"not null" json:"total_price"`
    Status     OrderStatus    `gorm:"type:varchar(20);default:pending" json:"status"`
    CreatedAt  time.Time      `json:"created_at"`
    UpdatedAt  time.Time      `json:"updated_at"`
    DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

type CreateOrderRequest struct {
    ProductID uint `json:"product_id" validate:"required"`
    Quantity  int  `json:"quantity" validate:"required,gt=0"`
}

// OrderEvent adalah pesan yang dikirim ke Kafka
type OrderEvent struct {
    OrderID   uint        `json:"order_id"`
    ProductID uint        `json:"product_id"`
    Quantity  int         `json:"quantity"`
    Status    OrderStatus `json:"status"`
}