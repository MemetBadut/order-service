package grpcclient

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MemetBadut/order-service/internal/resilience"
	pb "github.com/MemetBadut/order-service/proto/product"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// ProductClient adalah wrapper untuk gRPC client ke Product Service
type ProductClient struct {
	client  pb.ProductServiceClient
	conn    *grpc.ClientConn
	breaker *resilience.CircuitBreaker
}

type ServiceUnavailableError struct {
	Service string
	Cause   error
}

func (e *ServiceUnavailableError) Error() string {
	return fmt.Sprintf("%s sedang tidak tersedia: %v", e.Service, e.Cause)
}

func (e *ServiceUnavailableError) Unwrap() error {
	return e.Cause
}

func IsServiceUnavailable(err error) bool {
	var target *ServiceUnavailableError
	return errors.As(err, &target)
}

func NewProductClient(address string) (*ProductClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("gagal koneksi ke product service: %w", err)
	}

	return &ProductClient{
		client: pb.NewProductServiceClient(conn),
		conn:   conn,
		breaker: resilience.NewCircuitBreaker(resilience.Config{
			MaxFailures:  3,
			ResetTimeout: 10 * time.Second,
			IsFailure:    isRetryableGRPCError,
		}),
	}, nil
}

func (c *ProductClient) Close() {
	c.conn.Close()
}

func (c *ProductClient) CheckStock(productID uint64, quantity int32) (bool, string, error) {
	var resp *pb.CheckStockResponse

	err := c.breaker.Execute(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var callErr error
		resp, callErr = c.client.CheckStock(ctx, &pb.CheckStockRequest{
			ProductId: productID,
			Quantity:  quantity,
		})
		return callErr
	})
	if err != nil {
		return false, "", wrapProductServiceError("CheckStock", err)
	}

	return resp.Available, resp.Message, nil
}

func (c *ProductClient) GetProduct(productID uint64) (*pb.GetProductResponse, error) {
	var resp *pb.GetProductResponse

	err := c.breaker.Execute(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var callErr error
		resp, callErr = c.client.GetProduct(ctx, &pb.GetProductRequest{ProductId: productID})
		return callErr
	})
	if err != nil {
		return nil, wrapProductServiceError("GetProduct", err)
	}

	return resp, nil
}

func wrapProductServiceError(operation string, err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, resilience.ErrCircuitOpen) || isRetryableGRPCError(err) {
		return &ServiceUnavailableError{
			Service: "product-service",
			Cause:   fmt.Errorf("%s gagal: %w", operation, err),
		}
	}

	return fmt.Errorf("%s gagal: %w", operation, err)
}

func isRetryableGRPCError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	switch status.Code(err) {
	case codes.DeadlineExceeded, codes.Unavailable, codes.Internal, codes.ResourceExhausted, codes.Unknown:
		return true
	default:
		return false
	}
}
