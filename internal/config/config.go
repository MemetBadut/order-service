package config

import(
	"fmt"
	"os"


	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct{
	AppPort string
	GRPCPort string
	DB *gorm.DB
}

func Load() (*Config, error){
	_= godotenv.Load()

    dsn := fmt.Sprintf(
        "host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
        os.Getenv("DB_HOST"),
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_NAME"),
        os.Getenv("DB_PORT"),
        os.Getenv("DB_SSLMODE"),
    )

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("Gagal Koneksi Database: %w", err)
	}

	return &Config{
		AppPort: os.Getenv("APP_PORT"),
		GRPCPort: os.Getenv("GRPC_PORT"),
		DB: db,
	}, nil
}