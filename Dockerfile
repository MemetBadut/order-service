# Stage 1: Build
FROM golang:1.26.1-alpine AS builder

# Install git untuk submodules
RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod dan go.sum dulu (cache layer untuk dependency)
COPY go.mod go.sum ./
RUN go mod download

# Copy semua source code termasuk submodule proto
COPY . .

# Build binary (CGO disabled untuk static binary di Alpine)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server/main.go

# Stage 2: Runtime
# Gunakan image yang sangat kecil
FROM alpine:3.19

# Install ca-certificates untuk HTTPS calls
RUN apk --no-cache add ca-certificates tzdata

# Set timezone ke WIB
ENV TZ=Asia/Jakarta

WORKDIR /root/

# Copy binary dari stage build
COPY --from=builder /app/main .

# Expose port (dokumentasi saja, tidak benar-benar membuka port)
EXPOSE 8081

# Jalankan binary
CMD ["./main"]