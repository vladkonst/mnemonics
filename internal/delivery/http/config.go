package http

import (
	"os"
	"path/filepath"
)

// Config holds all configuration for the HTTP server and its dependencies.
// Values are read from environment variables with defaults.
type Config struct {
	Addr       string
	AdminToken string
	DBPath     string
	UploadsDir string
	LogLevel   string
	LogFormat  string

	S3Endpoint  string
	S3Bucket    string
	S3AccessKey string
	S3SecretKey string
	S3Region    string

	PaymentURL    string
	PaymentShopID string
	PaymentSecret string
	WebhookSecret string
}

// LoadConfig reads configuration from environment variables, applying defaults where needed.
func LoadConfig() Config {
	dbPath := getEnvOrDefault("SQLITE_PATH", getEnvOrDefault("DB_PATH", "./mnemo.db"))
	return Config{
		Addr:       getEnvOrDefault("SERVER_ADDR", ":8080"),
		AdminToken: getEnvOrDefault("ADMIN_TOKEN", "changeme"),
		DBPath:     dbPath,
		UploadsDir: getEnvOrDefault("UPLOADS_DIR", filepath.Join(filepath.Dir(dbPath), "uploads")),
		LogLevel:   getEnvOrDefault("LOG_LEVEL", "info"),
		LogFormat:  getEnvOrDefault("LOG_FORMAT", "console"),

		S3Endpoint:  os.Getenv("S3_ENDPOINT"),
		S3Bucket:    os.Getenv("S3_BUCKET"),
		S3AccessKey: os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey: os.Getenv("S3_SECRET_KEY"),
		S3Region:    getEnvOrDefault("S3_REGION", "us-east-1"),

		PaymentURL:    os.Getenv("PAYMENT_GATEWAY_URL"),
		PaymentShopID: os.Getenv("PAYMENT_SHOP_ID"),
		PaymentSecret: os.Getenv("PAYMENT_SECRET_KEY"),
		WebhookSecret: os.Getenv("PAYMENT_WEBHOOK_SECRET"),
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
