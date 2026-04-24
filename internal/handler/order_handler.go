package handler

import (
	"strconv"

	grpcclient "github.com/MemetBadut/order-service/internal/grpc"
	"github.com/MemetBadut/order-service/internal/model"
	"github.com/MemetBadut/order-service/internal/service"
	"github.com/gofiber/fiber/v3"
)

type OrderHandler struct {
	svc service.OrderService
}

func NewOrderHandler(svc service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) RegisterRoutes(router fiber.Router) {
	orders := router.Group("/orders")
	orders.Get("/", h.GetAll)
	orders.Get("/:id", h.GetByID)
	orders.Post("/", h.Create)
}

func (h *OrderHandler) GetAll(c fiber.Ctx) error {
	orders, err := h.svc.GetAllOrders()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": orders})
}

func (h *OrderHandler) GetByID(c fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 32)
	order, err := h.svc.GetOrderByID(uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": order})
}

func (h *OrderHandler) Create(c fiber.Ctx) error {
	var req model.CreateOrderRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "request tidak valid"})
	}
	order, err := h.svc.CreateOrder(&req)
	if err != nil {
		if grpcclient.IsServiceUnavailable(err) {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": err.Error()})
		}

		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"data": order, "message": "order berhasil dibuat"})
}
