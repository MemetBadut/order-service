package service

import (
	"errors"

	grpcclient "github.com/MemetBadut/order-service/internal/grpc"
	kafkapkg "github.com/MemetBadut/order-service/internal/kafka"
	"github.com/MemetBadut/order-service/internal/model"
	"github.com/MemetBadut/order-service/internal/repository"
)

type orderService struct {
	repo          repository.OrderRepository
	productClient *grpcclient.ProductClient
	producer      *kafkapkg.OrderProducer // BARU: Kafka producer
}

type OrderService interface {
	GetAllOrders() ([]model.Order, error)
	GetOrderByID(id uint) (*model.Order, error)
	CreateOrder(req *model.CreateOrderRequest) (*model.Order, error)
}

func NewOrderService(
	repo repository.OrderRepository,
	pc *grpcclient.ProductClient,
	producer *kafkapkg.OrderProducer,
) OrderService {
	return &orderService{repo: repo, productClient: pc, producer: producer}
}

func (s *orderService) GetAllOrders() ([]model.Order, error) {
	return s.repo.FindAll()
}

func (s *orderService) GetOrderByID(id uint) (*model.Order, error) {
	return s.repo.FindByID(id)
}

func (s *orderService) CreateOrder(req *model.CreateOrderRequest) (*model.Order, error) {
	// 1. Cek stok via gRPC
	available, msg, err := s.productClient.CheckStock(uint64(req.ProductID), int32(req.Quantity))
	if err != nil || !available {
		if err == nil {
			err = errors.New(msg)
		}
		return nil, err
	}

	// 2. Ambil harga produk
	productDetail, err := s.productClient.GetProduct(uint64(req.ProductID))
	if err != nil {
		return nil, errors.New("gagal ambil detail produk")
	}
	// 3. Buat order
	// Karena stok sudah lolos validasi sinkron via gRPC,
	// order langsung disimpan sebagai confirmed.
	order := &model.Order{
		ProductID:  req.ProductID,
		Quantity:   req.Quantity,
		TotalPrice: productDetail.Price * float64(req.Quantity),
		Status:     model.StatusConfirmed,
	}

	if err := s.repo.Create(order); err != nil {
		return nil, errors.New("gagal simpan order")
	}

	// 4. Publish event order.created ke Kafka setelah order tersimpan
	event := kafkapkg.OrderEventPayload{
		EventName: "order.created",
		OrderID:   order.ID,
		ProductID: order.ProductID,
		Quantity:  order.Quantity,
		Status:    string(order.Status),
	}

	// Publish di goroutine agar tidak blocking response
	if s.producer != nil {
		go func() {
			if err := s.producer.PublishOrderCreated(event); err != nil {
				// Log error tapi jangan gagalkan order
				// Di production: retry logic atau dead letter queue
				// `_ = err` artinya nilai error sengaja diabaikan setelah titik ini,
				// supaya compiler tahu perilaku ini memang disengaja.
				_ = err
			}
		}()
	}
	return order, nil
}
