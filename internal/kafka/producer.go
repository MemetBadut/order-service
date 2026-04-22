package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/segmentio/kafka-go"
	// "github.com/MemetBadut/order-service/internal/model"
)

const TopicOrderEvents = "order-events"

type OrderProducer struct {
	writer *kafka.Writer
}

type OrderEventPayload struct {
	EventName string `json:"event_name"`
	OrderID   uint   `json:"order_id"`
	ProductID uint   `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Status    string `json:"status"`
}

func NewOrderProducer(brokerAddr string) *OrderProducer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokerAddr),
		Topic:    TopicOrderEvents,
		Balancer: &kafka.LeastBytes{},
		// Auto-create topic jika belum ada
		AllowAutoTopicCreation: true,
	}
	return &OrderProducer{writer: writer}
}

func (p *OrderProducer) PublishOrderCreated(event OrderEventPayload) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("gagal marshal event: %w", err)
	}

	msg := kafka.Message{
		// Key digunakan untuk partitioning
		// (semua event order yang sama ke partisi yang sama)
		Key:   []byte(strconv.FormatUint(uint64(event.OrderID), 10)),
		Value: payload,
	}

	if err := p.writer.WriteMessages(context.Background(), msg); err != nil {
		return fmt.Errorf("gagal publish ke kafka: %w", err)
	}

	log.Printf("Event published: %s OrderID=%d, ProductID=%d, Qty=%d, Status=%s",
		event.EventName, event.OrderID, event.ProductID, event.Quantity, event.Status)

	return nil
}

func (p *OrderProducer) Close() {
	p.writer.Close()
}
