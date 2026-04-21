package grpcclient

import (
    "context"
    "fmt"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    pb "github.com/Memetbadut/order-service/proto/product"
)

// ProductClient adalah wrapper untuk gRPC client ke Product Service
type ProductClient struct {
    client pb.ProductServiceClient
    conn   *grpc.ClientConn
}

func NewProductClient(address string) (*ProductClient, error) {
    // Buat koneksi gRPC (insecure untuk development)
    // Untuk production: gunakan TLS credentials
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
    }, nil
}

// Close menutup koneksi gRPC (panggil saat service shutdown)
func (c *ProductClient) Close() {
    c.conn.Close()
}

// CheckStock memanggil Product Service untuk cek ketersediaan stok
func (c *ProductClient) CheckStock(productID uint64, quantity int32) (bool, string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    resp, err := c.client.CheckStock(ctx, &pb.CheckStockRequest{
        ProductId: productID,
        Quantity:  quantity,
    })
    if err != nil {
        return false, "", fmt.Errorf("gRPC CheckStock error: %w", err)
    }

    return resp.Available, resp.Message, nil
}

// GetProduct mengambil detail produk dari Product Service
func (c *ProductClient) GetProduct(productID uint64) (*pb.GetProductResponse, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    return c.client.GetProduct(ctx, &pb.GetProductRequest{ProductId: productID})
}