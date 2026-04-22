package main

import (
	"log"
	"os"

	"github.com/MemetBadut/order-service/internal/config"
	grpcclient "github.com/MemetBadut/order-service/internal/grpc"
	"github.com/MemetBadut/order-service/internal/handler"
	"github.com/MemetBadut/order-service/internal/kafka"
	"github.com/MemetBadut/order-service/internal/model"
	"github.com/MemetBadut/order-service/internal/repository"
	"github.com/MemetBadut/order-service/internal/service"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	cfg.DB.AutoMigrate(&model.Order{})

	// Setup gRPC client ke Product Service
	productGRPCAddr := os.Getenv("PRODUCT_GRPC_ADDR") // misal: "product-service:9091"
	productClient, err := grpcclient.NewProductClient(productGRPCAddr)
	if err != nil {
		log.Fatalf("Gagal koneksi gRPC ke product service: %v", err)
	}
	defer productClient.Close()

	kafkarBroker := os.Getenv("KAFKA_BROKER")
	orderProducer := kafka.NewOrderProducer(kafkarBroker)
	defer orderProducer.Close()

	orderRepo := repository.NewOrderRepository(cfg.DB)
	orderSvc := service.NewOrderService(orderRepo, productClient, orderProducer)
	orderHandler := handler.NewOrderHandler(orderSvc)

	app := fiber.New(fiber.Config{AppName: "Order Service v1.0"})
	app.Use(logger.New())

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "order-service"})
	})

	api := app.Group("/api/v1")
	orderHandler.RegisterRoutes(api)

	log.Printf("Order Service berjalan di port %s", cfg.AppPort)
	log.Fatal(app.Listen(":" + cfg.AppPort))
}
