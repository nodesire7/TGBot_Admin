package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// SetupConfig represents the configuration for setup wizard
type SetupConfig struct {
	// Database
	DBHost     string `json:"db_host"`
	DBPort     string `json:"db_port"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
	DBName     string `json:"db_name"`

	// Redis
	RedisHost     string `json:"redis_host"`
	RedisPort     string `json:"redis_port"`
	RedisPassword string `json:"redis_password"`

	// Admin
	AdminUsername string `json:"admin_username"`
	AdminPassword string `json:"admin_password"`
	JWTSecret     string `json:"jwt_secret"`

	// Bot
	BotToken string `json:"bot_token"`

	// Status
	IsConfigured bool `json:"is_configured"`
}

var (
	setupConfig     *SetupConfig
	setupConfigLock sync.RWMutex
	configFilePath  = "/app/data/config.json"
)

// GetConfigFilePath returns the config file path
func GetConfigFilePath() string {
	return configFilePath
}

// SetConfigFilePath sets a custom config file path (for testing)
func SetConfigFilePath(path string) {
	configFilePath = path
}

// LoadSetupConfig loads configuration from file
func LoadSetupConfig() (*SetupConfig, error) {
	setupConfigLock.Lock()
	defer setupConfigLock.Unlock()

	// Try to load from file
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config
			setupConfig = &SetupConfig{
				DBHost:        getEnvOrDefault("DB_HOST", "postgres"),
				DBPort:        getEnvOrDefault("DB_PORT", "5432"),
				DBUser:        getEnvOrDefault("DB_USER", "tgbot"),
				DBPassword:    getEnvOrDefault("DB_PASSWORD", "tgbot123"),
				DBName:        getEnvOrDefault("DB_NAME", "tgbot"),
				RedisHost:     getEnvOrDefault("REDIS_HOST", "redis"),
				RedisPort:     getEnvOrDefault("REDIS_PORT", "6379"),
				RedisPassword: getEnvOrDefault("REDIS_PASSWORD", ""),
				AdminUsername: getEnvOrDefault("ADMIN_USERNAME", "admin"),
				AdminPassword: getEnvOrDefault("ADMIN_PASSWORD", "admin123"),
				JWTSecret:     getEnvOrDefault("JWT_SECRET", ""),
				BotToken:      getEnvOrDefault("BOT_TOKEN", ""),
				IsConfigured:  false,
			}
			return setupConfig, nil
		}
		return nil, err
	}

	setupConfig = &SetupConfig{}
	if err := json.Unmarshal(data, setupConfig); err != nil {
		return nil, err
	}

	return setupConfig, nil
}

// SaveSetupConfig saves configuration to file
func SaveSetupConfig(cfg *SetupConfig) error {
	setupConfigLock.Lock()
	defer setupConfigLock.Unlock()

	// Ensure directory exists
	dir := filepath.Dir(configFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	cfg.IsConfigured = true
	setupConfig = cfg

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFilePath, data, 0600)
}

// GetSetupConfig returns current setup config
func GetSetupConfig() *SetupConfig {
	setupConfigLock.RLock()
	defer setupConfigLock.RUnlock()
	return setupConfig
}

// IsConfigured returns true if system has been configured
func IsConfigured() bool {
	setupConfigLock.RLock()
	defer setupConfigLock.RUnlock()
	return setupConfig != nil && setupConfig.IsConfigured
}

// TestDatabaseConnection tests database connectivity
func TestDatabaseConnection(cfg *SetupConfig) error {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return fmt.Errorf("failed to create connection: %v", err)
	}
	defer pool.Close()

	return pool.Ping(ctx)
}

// TestRedisConnection tests Redis connectivity
func TestRedisConnection(cfg *SetupConfig) error {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       0,
	})
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return client.Ping(ctx).Err()
}

// InitDBFromConfig initializes database from setup config
func InitDBFromConfig() error {
	cfg := GetSetupConfig()
	if cfg == nil {
		return fmt.Errorf("config not loaded")
	}

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	var err error
	dbPool, err = pgxpool.New(context.Background(), connStr)
	if err != nil {
		return err
	}

	dbPool.Config().MaxConns = 25
	dbPool.Config().MinConns = 5
	dbPool.Config().MaxConnLifetime = time.Hour
	dbPool.Config().MaxConnIdleTime = 30 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return dbPool.Ping(ctx)
}

// InitRedisFromConfig initializes Redis from setup config
func InitRedisFromConfig() error {
	cfg := GetSetupConfig()
	if cfg == nil {
		return fmt.Errorf("config not loaded")
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       0,
		PoolSize: 50,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return redisClient.Ping(ctx).Err()
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
