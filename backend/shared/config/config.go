package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type DatabaseConfig struct {
	URL                   string
	Host                  string
	Port                  int
	Name                  string
	User                  string
	Password              string
	SSLMode               string
	MaxConns              int
	MinConns              int
	ConnTimeoutSeconds    int
	QueryTimeoutSeconds   int
	RDSProxyEndpoint      string
	RDSIAMAuthEnabled     bool
}

type RedisClusterConfig struct {
	Host     string
	Port     int
	Password string
}

type RedisConfig struct {
	Cart         RedisClusterConfig
	Session      RedisClusterConfig
	SKUCache     RedisClusterConfig
	RateLimit    RedisClusterConfig
	ExitToken    RedisClusterConfig
	Realtime     RedisClusterConfig
	ConnTimeoutMS int
}

type KafkaConfig struct {
	Brokers                  string
	SecurityProtocol         string
	SASLMechanism            string
	SASLUsername             string
	SASLPassword             string
	ConsumerGroupPrefix      string
	SchemaRegistryURL        string
	TopicPartitionsDefault   int
	TopicPartitionsCartScan  int
}

type ESConfig struct {
	Host           string
	Port           int
	Username       string
	Password       string
	UseTLS         bool
	IndexPrefix    string
	QueryTimeoutMS int
}

type ClickHouseConfig struct {
	Host     string
	Port     int
	HTTPPort int
	DB       string
	User     string
	Password string
}

type TimescaleConfig struct {
	URL string
}

type JWTConfig struct {
	PrivateKeyCurrent    string
	PublicKeyCurrent     string
	KIDCurrent           string
	PrivateKeyPrevious   string
	PublicKeyPrevious    string
	KIDPrevious          string
	AccessTTLMinutes     int
	RefreshTTLDays       int
	ExitTokenTTLMinutes  int
	Issuer               string
	KeyRotationDays      int
	Secret               string
}

type OTPConfig struct {
	Length                       int
	TTLSeconds                   int
	MaxWrongAttempts             int
	LockoutMinutes               int
	ResendCooldownSeconds        int
	RateLimitPerIdentifier15Min  int
	RateLimitPerIP15Min          int

	SMSProvider             string
	TwilioAccountSID        string
	TwilioAuthToken         string
	TwilioVerifyServiceSID  string
	MSG91AuthKey            string
	MSG91TemplateID         string
	MSG91SenderID           string
	MSG91PromotionalSenderID string

	GmailSMTPHost        string
	GmailSMTPPort        int
	GmailSMTPUser        string
	GmailSMTPAppPassword string
	EmailFromName        string

	GoogleOAuthClientIDWeb     string
	GoogleOAuthClientIDAndroid string
	GoogleOAuthClientIDiOS     string
	GoogleOAuthClientSecret    string
}

type PaymentConfig struct {
	PrimaryGateway                   string
	RazorpayKeyID                    string
	RazorpayKeySecret                string
	RazorpayWebhookSecret            string
	RazorpayWebhookIPAllowlist      string
	CashfreeAppID                    string
	CashfreeSecretKey                string
	PayUMerchantKey                  string
	PayUSalt                         string
	OutboundTimeoutSeconds           int
	CircuitBreakerErrorThreshold     float64
}

type NotificationsConfig struct {
	WhatsAppBusinessAccountID string
	WhatsAppPhoneNumberID     string
	WhatsAppAccessToken       string
	WhatsAppWebhookVerifyToken string

	FirebaseProjectID              string
	FirebaseServiceAccountJSONPath string
}

type AWSConfig struct {
	AccessKeyID                 string
	SecretAccessKey             string
	S3BucketInvoices            string
	S3BucketProducts            string
	S3BucketMedia               string
	SecretsManagerPrefix        string
	SecretsRotationPollMinutes  int
	CloudFrontDomain            string
	CloudFrontKeyPairID         string
	CloudFrontPrivateKeyPath    string
	IoTEndpoint                 string
	IoTCertPath                 string
	IoTKeyPath                  string
	IoTCAPath                   string
	MQTTTopicPrefix             string
}

type ComplianceConfig struct {
	GSTIRPBaseURL             string
	GSTIRPGSPUsername         string
	GSTIRPGSPPassword         string
	GSTIRPClientID            string
	GSTIRPClientSecret        string
	GSTNTaxpayerAPIKey        string
	DPDPGrievanceOfficerEmail string
	DPDPDataRetentionDays     int
}

type SecurityConfig struct {
	AllowedCORSOrigins            string
	MaxRequestBodyBytes           int64
	GooglePlayIntegrityPackageName string
	GooglePlayIntegrityDecodeURL   string
	WAFEnabled                     bool
}

type ObservabilityConfig struct {
	OTELExporterOTLPEndpoint string
	OTELServiceNamePrefix    string
	JaegerEndpoint           string
	PrometheusPushgateway    string
	SentryDSNBackend         string
	PagerDutyIntegrationKey  string
	StatusPageAPIKey         string
}

type FeatureFlagsConfig struct {
	RedisPrefix    string
	CacheTTLSeconds int
}

type ERPConfig struct {
	WebhookSharedSecret string
	SAPAPIBaseURL       string
	TallyAPIBaseURL     string
}

type Config struct {
	ServiceName string
	Port        string
	Environment string
	AppVersion  string
	LogLevel    string
	AWSRegion   string

	Database      DatabaseConfig
	Redis         RedisConfig
	Kafka         KafkaConfig
	ES            ESConfig
	ClickHouse    ClickHouseConfig
	Timescale     TimescaleConfig
	JWT           JWTConfig
	OTP           OTPConfig
	Payment       PaymentConfig
	Notifications NotificationsConfig
	AWS           AWSConfig
	Compliance    ComplianceConfig
	Security      SecurityConfig
	Observability ObservabilityConfig
	FeatureFlags  FeatureFlagsConfig
	ERP           ERPConfig
}

func Load(serviceName string) (*Config, error) {
	env := getEnv("ENVIRONMENT", "development")

	// 1. In dev, load .env file if present
	if env == "development" {
		_ = godotenv.Load(".env")
		_ = godotenv.Load("../.env")
		_ = godotenv.Load("../../.env")
	}

	cfg := &Config{
		ServiceName: serviceName,
		Port:        getEnv("PORT", getServiceDefaultPort(serviceName)),
		Environment: env,
		AppVersion:  getEnv("APP_VERSION", "1.0.0"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		AWSRegion:   getEnv("AWS_REGION", "ap-south-1"),

		Database: DatabaseConfig{
			URL:                 getEnv("DATABASE_URL", "postgres://zippyra:changeme@localhost:5432/zippyra?sslmode=disable"),
			Host:                getEnv("DB_HOST", "localhost"),
			Port:                getEnvInt("DB_PORT", 5432),
			Name:                getEnv("DB_NAME", "zippyra"),
			User:                getEnv("DB_USER", "zippyra"),
			Password:            getEnv("DB_PASSWORD", "changeme"),
			SSLMode:             getEnv("DB_SSL_MODE", "disable"),
			MaxConns:            getEnvInt("DB_MAX_CONNS", 10),
			MinConns:            getEnvInt("DB_MIN_CONNS", 2),
			ConnTimeoutSeconds:  getEnvInt("DB_CONN_TIMEOUT_SECONDS", 5),
			QueryTimeoutSeconds: getEnvInt("DB_QUERY_TIMEOUT_SECONDS", 5),
			RDSProxyEndpoint:    getEnv("RDS_PROXY_ENDPOINT", ""),
			RDSIAMAuthEnabled:   getEnvBool("RDS_IAM_AUTH_ENABLED", false),
		},

		Redis: RedisConfig{
			Cart:          RedisClusterConfig{Host: getEnv("REDIS_CART_HOST", "localhost"), Port: getEnvInt("REDIS_CART_PORT", 6379), Password: getEnv("REDIS_CART_PASSWORD", "")},
			Session:       RedisClusterConfig{Host: getEnv("REDIS_SESSION_HOST", "localhost"), Port: getEnvInt("REDIS_SESSION_PORT", 6379), Password: getEnv("REDIS_SESSION_PASSWORD", "")},
			SKUCache:      RedisClusterConfig{Host: getEnv("REDIS_SKU_CACHE_HOST", "localhost"), Port: getEnvInt("REDIS_SKU_CACHE_PORT", 6379), Password: getEnv("REDIS_SKU_CACHE_PASSWORD", "")},
			RateLimit:     RedisClusterConfig{Host: getEnv("REDIS_RATE_LIMIT_HOST", "localhost"), Port: getEnvInt("REDIS_RATE_LIMIT_PORT", 6379), Password: getEnv("REDIS_RATE_LIMIT_PASSWORD", "")},
			ExitToken:     RedisClusterConfig{Host: getEnv("REDIS_EXIT_TOKEN_HOST", "localhost"), Port: getEnvInt("REDIS_EXIT_TOKEN_PORT", 6379), Password: getEnv("REDIS_EXIT_TOKEN_PASSWORD", "")},
			Realtime:      RedisClusterConfig{Host: getEnv("REDIS_REALTIME_HOST", "localhost"), Port: getEnvInt("REDIS_REALTIME_PORT", 6379), Password: getEnv("REDIS_REALTIME_PASSWORD", "")},
			ConnTimeoutMS: getEnvInt("REDIS_CONN_TIMEOUT_MS", 100),
		},

		Kafka: KafkaConfig{
			Brokers:                 getEnv("KAFKA_BROKERS", "localhost:9092"),
			SecurityProtocol:        getEnv("KAFKA_SECURITY_PROTOCOL", "PLAINTEXT"),
			SASLMechanism:           getEnv("KAFKA_SASL_MECHANISM", ""),
			SASLUsername:            getEnv("KAFKA_SASL_USERNAME", ""),
			SASLPassword:            getEnv("KAFKA_SASL_PASSWORD", ""),
			ConsumerGroupPrefix:     getEnv("KAFKA_CONSUMER_GROUP_PREFIX", "zippyra-dev"),
			SchemaRegistryURL:       getEnv("KAFKA_SCHEMA_REGISTRY_URL", ""),
			TopicPartitionsDefault:  getEnvInt("KAFKA_TOPIC_PARTITIONS_DEFAULT", 6),
			TopicPartitionsCartScan: getEnvInt("KAFKA_TOPIC_PARTITIONS_CART_SCAN", 256),
		},

		ES: ESConfig{
			Host:           getEnv("ES_HOST", "localhost"),
			Port:           getEnvInt("ES_PORT", 9200),
			Username:       getEnv("ES_USERNAME", ""),
			Password:       getEnv("ES_PASSWORD", ""),
			UseTLS:         getEnvBool("ES_USE_TLS", false),
			IndexPrefix:    getEnv("ES_INDEX_PREFIX", "zippyra_dev"),
			QueryTimeoutMS: getEnvInt("ES_QUERY_TIMEOUT_MS", 800),
		},

		ClickHouse: ClickHouseConfig{
			Host:     getEnv("CLICKHOUSE_HOST", "localhost"),
			Port:     getEnvInt("CLICKHOUSE_PORT", 9000),
			HTTPPort: getEnvInt("CLICKHOUSE_HTTP_PORT", 8123),
			DB:       getEnv("CLICKHOUSE_DB", "zippyra_analytics"),
			User:     getEnv("CLICKHOUSE_USER", "default"),
			Password: getEnv("CLICKHOUSE_PASSWORD", ""),
		},

		Timescale: TimescaleConfig{
			URL: getEnv("TIMESCALE_URL", "postgres://zippyra:changeme@localhost:5433/zippyra_timeseries?sslmode=disable"),
		},

		JWT: JWTConfig{
			PrivateKeyCurrent:   getEnv("JWT_PRIVATE_KEY_CURRENT", ""),
			PublicKeyCurrent:    getEnv("JWT_PUBLIC_KEY_CURRENT", ""),
			KIDCurrent:          getEnv("JWT_KID_CURRENT", "key-2026-07-a"),
			PrivateKeyPrevious:  getEnv("JWT_PRIVATE_KEY_PREVIOUS", ""),
			PublicKeyPrevious:   getEnv("JWT_PUBLIC_KEY_PREVIOUS", ""),
			KIDPrevious:         getEnv("JWT_KID_PREVIOUS", ""),
			AccessTTLMinutes:    getEnvInt("JWT_ACCESS_TTL_MINUTES", 15),
			RefreshTTLDays:      getEnvInt("JWT_REFRESH_TTL_DAYS", 30),
			ExitTokenTTLMinutes: getEnvInt("JWT_EXIT_TOKEN_TTL_MINUTES", 10),
			Issuer:              getEnv("JWT_ISSUER", "zippyra.com"),
			KeyRotationDays:     getEnvInt("JWT_KEY_ROTATION_DAYS", 90),
			Secret:              getEnv("JWT_SECRET", "zippyra-dev-jwt-secret-key-32bytes"),
		},

		OTP: OTPConfig{
			Length:                      getEnvInt("OTP_LENGTH", 6),
			TTLSeconds:                  getEnvInt("OTP_TTL_SECONDS", 300),
			MaxWrongAttempts:            getEnvInt("OTP_MAX_WRONG_ATTEMPTS", 5),
			LockoutMinutes:              getEnvInt("OTP_LOCKOUT_MINUTES", 15),
			ResendCooldownSeconds:       getEnvInt("OTP_RESEND_COOLDOWN_SECONDS", 30),
			RateLimitPerIdentifier15Min: getEnvInt("OTP_RATE_LIMIT_PER_IDENTIFIER_PER_15MIN", 5),
			RateLimitPerIP15Min:         getEnvInt("OTP_RATE_LIMIT_PER_IP_PER_15MIN", 10),

			SMSProvider:            getEnv("SMS_PROVIDER", "twilio"),
			TwilioAccountSID:       getEnv("TWILIO_ACCOUNT_SID", ""),
			TwilioAuthToken:        getEnv("TWILIO_AUTH_TOKEN", ""),
			TwilioVerifyServiceSID: getEnv("TWILIO_VERIFY_SERVICE_SID", ""),
			MSG91AuthKey:           getEnv("MSG91_AUTH_KEY", ""),
			MSG91TemplateID:        getEnv("MSG91_TEMPLATE_ID", ""),
			MSG91SenderID:          getEnv("MSG91_SENDER_ID", "ZPYRA"),
			MSG91PromotionalSenderID: getEnv("MSG91_PROMOTIONAL_SENDER_ID", "PRMZIP"),

			GmailSMTPHost:        getEnv("GMAIL_SMTP_HOST", "smtp.gmail.com"),
			GmailSMTPPort:        getEnvInt("GMAIL_SMTP_PORT", 587),
			GmailSMTPUser:        getEnv("GMAIL_SMTP_USER", ""),
			GmailSMTPAppPassword: getEnv("GMAIL_SMTP_APP_PASSWORD", ""),
			EmailFromName:        getEnv("EMAIL_FROM_NAME", "Zippyra"),

			GoogleOAuthClientIDWeb:     getEnv("GOOGLE_OAUTH_CLIENT_ID_WEB", getEnv("GOOGLE_OAUTH_CLIENT_ID", "")),
			GoogleOAuthClientIDAndroid: getEnv("GOOGLE_OAUTH_CLIENT_ID_ANDROID", ""),
			GoogleOAuthClientIDiOS:     getEnv("GOOGLE_OAUTH_CLIENT_ID_IOS", ""),
			GoogleOAuthClientSecret:    getEnv("GOOGLE_OAUTH_CLIENT_SECRET", ""),
		},

		Payment: PaymentConfig{
			PrimaryGateway:               getEnv("PAYMENT_PRIMARY_GATEWAY", "razorpay"),
			RazorpayKeyID:                getEnv("RAZORPAY_KEY_ID", ""),
			RazorpayKeySecret:            getEnv("RAZORPAY_KEY_SECRET", ""),
			RazorpayWebhookSecret:        getEnv("RAZORPAY_WEBHOOK_SECRET", ""),
			RazorpayWebhookIPAllowlist:  getEnv("RAZORPAY_WEBHOOK_IP_ALLOWLIST", ""),
			CashfreeAppID:                getEnv("CASHFREE_APP_ID", ""),
			CashfreeSecretKey:            getEnv("CASHFREE_SECRET_KEY", ""),
			PayUMerchantKey:              getEnv("PAYU_MERCHANT_KEY", ""),
			PayUSalt:                     getEnv("PAYU_SALT", ""),
			OutboundTimeoutSeconds:       getEnvInt("PAYMENT_OUTBOUND_TIMEOUT_SECONDS", 10),
			CircuitBreakerErrorThreshold: getEnvFloat("PAYMENT_CIRCUIT_BREAKER_ERROR_THRESHOLD", 0.05),
		},

		Notifications: NotificationsConfig{
			WhatsAppBusinessAccountID: getEnv("WHATSAPP_BUSINESS_ACCOUNT_ID", ""),
			WhatsAppPhoneNumberID:     getEnv("WHATSAPP_PHONE_NUMBER_ID", ""),
			WhatsAppAccessToken:       getEnv("WHATSAPP_ACCESS_TOKEN", ""),
			WhatsAppWebhookVerifyToken: getEnv("WHATSAPP_WEBHOOK_VERIFY_TOKEN", ""),

			FirebaseProjectID:              getEnv("FIREBASE_PROJECT_ID", ""),
			FirebaseServiceAccountJSONPath: getEnv("FIREBASE_SERVICE_ACCOUNT_JSON_PATH", "/secrets/firebase-service-account.json"),
		},

		AWS: AWSConfig{
			AccessKeyID:                getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey:            getEnv("AWS_SECRET_ACCESS_KEY", ""),
			S3BucketInvoices:           getEnv("AWS_S3_BUCKET_INVOICES", "zippyra-invoices-dev"),
			S3BucketProducts:           getEnv("AWS_S3_BUCKET_PRODUCTS", "zippyra-products-dev"),
			S3BucketMedia:              getEnv("AWS_S3_BUCKET_MEDIA", "zippyra-media-dev"),
			SecretsManagerPrefix:       getEnv("AWS_SECRETS_MANAGER_PREFIX", "zippyra/dev/"),
			SecretsRotationPollMinutes: getEnvInt("AWS_SECRETS_ROTATION_POLL_MINUTES", 5),
			CloudFrontDomain:           getEnv("CLOUDFRONT_DOMAIN", ""),
			CloudFrontKeyPairID:        getEnv("CLOUDFRONT_KEY_PAIR_ID", ""),
			CloudFrontPrivateKeyPath:   getEnv("CLOUDFRONT_PRIVATE_KEY_PATH", "/secrets/cloudfront-private-key.pem"),
			IoTEndpoint:                getEnv("AWS_IOT_ENDPOINT", ""),
			IoTCertPath:                getEnv("AWS_IOT_CERT_PATH", "/secrets/iot-device-cert.pem"),
			IoTKeyPath:                 getEnv("AWS_IOT_KEY_PATH", "/secrets/iot-device-key.pem"),
			IoTCAPath:                  getEnv("AWS_IOT_CA_PATH", "/secrets/iot-root-ca.pem"),
			MQTTTopicPrefix:            getEnv("MQTT_TOPIC_PREFIX", "zippyra"),
		},

		Compliance: ComplianceConfig{
			GSTIRPBaseURL:             getEnv("GST_IRP_BASE_URL", "https://einvoice1.gst.gov.in/eicore/v1.03"),
			GSTIRPGSPUsername:         getEnv("GST_IRP_GSP_USERNAME", ""),
			GSTIRPGSPPassword:         getEnv("GST_IRP_GSP_PASSWORD", ""),
			GSTIRPClientID:            getEnv("GST_IRP_CLIENT_ID", ""),
			GSTIRPClientSecret:        getEnv("GST_IRP_CLIENT_SECRET", ""),
			GSTNTaxpayerAPIKey:        getEnv("GSTN_TAXPAYER_API_KEY", ""),
			DPDPGrievanceOfficerEmail: getEnv("DPDP_GRIEVANCE_OFFICER_EMAIL", "grievance@zippyra.com"),
			DPDPDataRetentionDays:     getEnvInt("DPDP_DATA_RETENTION_DAYS", 730),
		},

		Security: SecurityConfig{
			AllowedCORSOrigins:            getEnv("ALLOWED_CORS_ORIGINS", "http://localhost:3000,http://localhost:3001,http://localhost:3002"),
			MaxRequestBodyBytes:           int64(getEnvInt("MAX_REQUEST_BODY_BYTES", 1048576)),
			GooglePlayIntegrityPackageName: getEnv("GOOGLE_PLAY_INTEGRITY_PACKAGE_NAME", "com.zippyra.customer"),
			GooglePlayIntegrityDecodeURL:   getEnv("GOOGLE_PLAY_INTEGRITY_DECODE_URL", "https://playintegrity.googleapis.com/v1"),
			WAFEnabled:                     getEnvBool("WAF_ENABLED", false),
		},

		Observability: ObservabilityConfig{
			OTELExporterOTLPEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317"),
			OTELServiceNamePrefix:    getEnv("OTEL_SERVICE_NAME_PREFIX", "zippyra"),
			JaegerEndpoint:           getEnv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
			PrometheusPushgateway:    getEnv("PROMETHEUS_PUSHGATEWAY", ""),
			SentryDSNBackend:         getEnv("SENTRY_DSN_BACKEND", ""),
			PagerDutyIntegrationKey:  getEnv("PAGERDUTY_INTEGRATION_KEY", ""),
			StatusPageAPIKey:         getEnv("STATUS_PAGE_API_KEY", ""),
		},

		FeatureFlags: FeatureFlagsConfig{
			RedisPrefix:    getEnv("FEATURE_FLAGS_REDIS_PREFIX", "feature:"),
			CacheTTLSeconds: getEnvInt("FEATURE_FLAGS_CACHE_TTL_SECONDS", 300),
		},

		ERP: ERPConfig{
			WebhookSharedSecret: getEnv("ERP_WEBHOOK_SHARED_SECRET", ""),
			SAPAPIBaseURL:       getEnv("SAP_API_BASE_URL", ""),
			TallyAPIBaseURL:     getEnv("TALLY_API_BASE_URL", ""),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	var missingVars []string

	// DPDP Data Localization Invariant: AWS_REGION must be ap-south-1 or ap-south-2 in every environment
	if c.AWSRegion != "ap-south-1" && c.AWSRegion != "ap-south-2" {
		missingVars = append(missingVars, fmt.Sprintf("AWS_REGION must be ap-south-1 or ap-south-2 for DPDP compliance (got %s)", c.AWSRegion))
	}

	// Service specific requirements
	switch c.ServiceName {
	case "auth-service":
		if c.Database.URL == "" {
			missingVars = append(missingVars, "DATABASE_URL")
		}
		if c.JWT.PrivateKeyCurrent == "" && c.JWT.Secret == "" {
			missingVars = append(missingVars, "JWT_PRIVATE_KEY_CURRENT or JWT_SECRET")
		}
		if c.JWT.KIDCurrent == "" {
			missingVars = append(missingVars, "JWT_KID_CURRENT")
		}
		if c.OTP.SMSProvider == "twilio" && (c.OTP.TwilioAccountSID == "" || c.OTP.TwilioAuthToken == "") {
			missingVars = append(missingVars, "TWILIO_ACCOUNT_SID / TWILIO_AUTH_TOKEN")
		}
		if (c.OTP.GmailSMTPUser != "" && c.OTP.GmailSMTPAppPassword == "") || (c.OTP.GmailSMTPUser == "" && c.OTP.GmailSMTPAppPassword != "") {
			missingVars = append(missingVars, "GMAIL_SMTP_USER and GMAIL_SMTP_APP_PASSWORD (must be provided together)")
		}

	case "payment-service":
		if c.Payment.RazorpayKeyID == "" {
			missingVars = append(missingVars, "RAZORPAY_KEY_ID")
		}
		if c.Payment.RazorpayKeySecret == "" {
			missingVars = append(missingVars, "RAZORPAY_KEY_SECRET")
		}
		if c.Payment.RazorpayWebhookSecret == "" {
			missingVars = append(missingVars, "RAZORPAY_WEBHOOK_SECRET")
		}
	}

	// Production Safety Invariants
	if c.Environment == "production" {
		if c.Database.SSLMode != "require" && c.Database.SSLMode != "verify-full" {
			missingVars = append(missingVars, fmt.Sprintf("PROD SAFETY: DB_SSL_MODE must be 'require' or 'verify-full' in production (got '%s')", c.Database.SSLMode))
		}
		if !c.Database.RDSIAMAuthEnabled {
			missingVars = append(missingVars, "PROD SAFETY: RDS_IAM_AUTH_ENABLED must be true in production")
		}
		if !c.Security.WAFEnabled {
			missingVars = append(missingVars, "PROD SAFETY: WAF_ENABLED must be true in production")
		}
		if c.Kafka.SecurityProtocol == "PLAINTEXT" {
			missingVars = append(missingVars, "PROD SAFETY: KAFKA_SECURITY_PROTOCOL must not be PLAINTEXT in production")
		}
	}

	if len(missingVars) > 0 {
		return fmt.Errorf("configuration validation failed for service '%s':\n- %s", c.ServiceName, strings.Join(missingVars, "\n- "))
	}

	return nil
}

func (c *Config) LogActiveIntegrations() {
	smsStatus := fmt.Sprintf("%s [active]", c.OTP.SMSProvider)
	emailStatus := "gmail-smtp [NOT CONFIGURED - stubbed]"
	if c.OTP.GmailSMTPUser != "" && c.OTP.GmailSMTPAppPassword != "" {
		emailStatus = "gmail-smtp [active]"
	}

	googleOAuthStatus := "[NOT CONFIGURED — endpoint returns 501]"
	if c.OTP.GoogleOAuthClientIDWeb != "" {
		googleOAuthStatus = "[active]"
	}

	fmt.Printf("[CONFIG] Service: %s | Env: %s | Region: %s | Version: %s\n", c.ServiceName, c.Environment, c.AWSRegion, c.AppVersion)
	fmt.Printf("[CONFIG] SMS Provider: %s | Email OTP: %s | Google OAuth: %s\n", smsStatus, emailStatus, googleOAuthStatus)
}

func getServiceDefaultPort(service string) string {
	ports := map[string]string{
		"auth-service":            "8080",
		"store-service":           "8082",
		"catalog-service":         "8083",
		"cart-service":            "8084",
		"payment-service":         "8085",
		"order-service":           "8086",
		"exit-validation-service": "8087",
		"notification-service":    "8088",
		"loyalty-service":         "8089",
		"inventory-service":       "8090",
		"warehouse-service":       "8091",
		"analytics-service":       "8092",
		"support-service":         "8093",
		"retailer-auth-service":   "8094",
		"admin-auth-service":      "8095",
		"chain-hq-service":        "8096",
		"admin-store-service":     "8097",
		"compliance-service":      "8098",
		"qc-service":              "8099",
		"transfer-service":        "8100",
		"audit-service":           "8101",
		"device-mgmt-service":     "8102",
		"integration-service":     "8103",
	}
	if port, ok := ports[service]; ok {
		return port
	}
	return "8080"
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if valStr, ok := os.LookupEnv(key); ok && valStr != "" {
		if val, err := strconv.Atoi(valStr); err == nil {
			return val
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if valStr, ok := os.LookupEnv(key); ok && valStr != "" {
		if val, err := strconv.ParseFloat(valStr, 64); err == nil {
			return val
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if valStr, ok := os.LookupEnv(key); ok && valStr != "" {
		if val, err := strconv.ParseBool(valStr); err == nil {
			return val
		}
	}
	return fallback
}
