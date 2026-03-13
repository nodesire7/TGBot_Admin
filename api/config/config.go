package config

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var (
	dbPool  *pgxpool.Pool
	redisClient *redis.Client
	once    sync.Once
)

// LoadEnv loads environment variables from .env file
func LoadEnv() error {
	return godotenv.Load()
}

// InitDB initializes PostgreSQL connection pool
func InitDB() error {
	var err error
	dbPool, err = pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}

	// Configure pool
	dbPool.Config().MaxConns = 25
	dbPool.Config().MinConns = 5
	dbPool.Config().MaxConnLifetime = time.Hour
	dbPool.Config().MaxConnIdleTime = 30 * time.Minute

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return dbPool.Ping(ctx)
}

// GetDB returns database pool instance
func GetDB() *pgxpool.Pool {
	return dbPool
}

// CloseDB closes database connection
func CloseDB() {
	if dbPool != nil {
		dbPool.Close()
	}
}

// InitRedis initializes Redis client
func InitRedis() error {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
		PoolSize: 50,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return redisClient.Ping(ctx).Err()
}

// GetRedis returns Redis client instance
func GetRedis() *redis.Client {
	return redisClient
}

// CloseRedis closes Redis connection
func CloseRedis() {
	if redisClient != nil {
		redisClient.Close()
	}
}

// GetJWTSecret returns JWT secret from env
func GetJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-change-in-production"
	}
	return []byte(secret)
}
