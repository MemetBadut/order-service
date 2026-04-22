package repository

import (
	"errors"

	"github.com/MemetBadut/order-service/internal/model"
	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(order *model.Order) error
	FindByID(id uint) (*model.Order, error)
	FindAll() ([]model.Order, error)
	UpdateStatus(id uint, status model.OrderStatus) error
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(order *model.Order) error {
	return r.db.Create(order).Error
}

func (r *orderRepository) FindByID(id uint) (*model.Order, error) {
	var order model.Order
	err := r.db.First(&order, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order tidak ditemukan")
		}
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindAll() ([]model.Order, error) {
	var orders []model.Order
	err := r.db.Find(&orders).Error
	return orders, err
}

func (r *orderRepository) UpdateStatus(id uint, status model.OrderStatus) error {
	return r.db.Model(&model.Order{}).Where("id = ?", id).
		Update("status", status).Error
}
