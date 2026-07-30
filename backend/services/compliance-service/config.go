package main

import "os"

type Config struct {
	Port              string
	DatabaseURL       string
	JWTSecret         string
	GSTIRPBaseURL     string
	GSTIRPUsername    string
	GSTIRPPassword    string
	GSTIRPClientID    string
	PaymentServiceURL string
	KafkaBrokers      string
}

func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8095"
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/zippyra_compliance?sslmode=disable"
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-key-change-in-prod"
	}
	paySvcURL := os.Getenv("PAYMENT_SERVICE_URL")
	if paySvcURL == "" {
		paySvcURL = "http://localhost:8083"
	}
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}

	return &Config{
		Port:              port,
		DatabaseURL:       dbURL,
		JWTSecret:         jwtSecret,
		GSTIRPBaseURL:     os.Getenv("GST_IRP_BASE_URL"),
		GSTIRPUsername:    os.Getenv("GST_IRP_USERNAME"),
		GSTIRPPassword:    os.Getenv("GST_IRP_PASSWORD"),
		GSTIRPClientID:    os.Getenv("GST_IRP_CLIENT_ID"),
		PaymentServiceURL: paySvcURL,
		KafkaBrokers:      kafkaBrokers,
	}
}
