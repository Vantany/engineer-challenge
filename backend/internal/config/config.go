package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBPort       int
	DBHost       string
	DBUser       string
	DBPassword   string
	DBName       string
	DBSSLMode    string
	JWTSecret    string
	PrivateKey   string
	PublicKey    string
	HTTPPort     string
	GRPCPort     string
	MetricsPort  string
	OTLPEndpoint string
}

func Load() *Config {
	dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))
	return &Config{
		DBPort:       dbPort,
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBUser:       getEnv("DB_USER", "auth"),
		DBPassword:   getEnv("DB_PASSWORD", "auth"),
		DBName:       getEnv("DB_NAME", "auth"),
		DBSSLMode:    getEnv("DB_SSLMODE", "disable"),
		JWTSecret:    getEnv("JWT_SECRET", "dev-secret-key-change-in-prod"),
		PrivateKey:   getEnv("JWT_PRIVATE_KEY", "keys/private.pem"),
		PublicKey:    getEnv("JWT_PUBLIC_KEY", "keys/public.pem"),
		HTTPPort:     getEnv("HTTP_PORT", "8080"),
		GRPCPort:     getEnv("GRPC_PORT", "9090"),
		MetricsPort:  getEnv("METRICS_PORT", "8081"),
		OTLPEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
