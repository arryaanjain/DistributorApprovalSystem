package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all runtime configuration for the API server.
// Values are loaded from environment variables (or a .env file).
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	OTP      OTPConfig
	Redis    RedisConfig
	Surepass SurepassConfig
	MSG91    MSG91Config
	Notify     NotifyConfig
	Razorpay   RazorpayConfig
	Shiprocket ShiprocketConfig
	Admin      AdminConfig
	App        AppConfig
}

type AdminConfig struct {
	Email    string `mapstructure:"ADMIN_EMAIL"`
	Password string `mapstructure:"ADMIN_PASSWORD"`
}

type RazorpayConfig struct {
	KeyID         string `mapstructure:"RAZORPAY_KEY_ID"`
	KeySecret     string `mapstructure:"RAZORPAY_KEY_SECRET"`
	WebhookSecret string `mapstructure:"RAZORPAY_WEBHOOK_SECRET"`
}

type ShiprocketConfig struct {
	Email          string `mapstructure:"SHIPROCKET_EMAIL"`
	Password       string `mapstructure:"SHIPROCKET_PASSWORD"`
	APIToken       string `mapstructure:"SHIPROCKET_API_TOKEN"`
	APIURL         string `mapstructure:"SHIPROCKET_API_URL"`
	ChannelID      string `mapstructure:"SHIPROCKET_CHANNEL_ID"`
	PickupLocation string `mapstructure:"SHIPROCKET_PICKUP_LOCATION"`
}

type ServerConfig struct {
	Port         string        `mapstructure:"PORT"`
	ReadTimeout  time.Duration `mapstructure:"READ_TIMEOUT"`
	WriteTimeout time.Duration `mapstructure:"WRITE_TIMEOUT"`
	IdleTimeout  time.Duration `mapstructure:"IDLE_TIMEOUT"`
	CORSOrigins  []string      `mapstructure:"CORS_ORIGINS"`
}

type DatabaseConfig struct {
	DSN             string `mapstructure:"DATABASE_URL"`
	MaxOpenConns    int    `mapstructure:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns    int    `mapstructure:"DB_MAX_IDLE_CONNS"`
	ConnMaxLifetime time.Duration
	MigrationsDir   string `mapstructure:"MIGRATIONS_DIR"`
}

type JWTConfig struct {
	AccessSecret      string        `mapstructure:"JWT_ACCESS_SECRET"`
	RefreshSecret     string        `mapstructure:"JWT_REFRESH_SECRET"`
	AccessExpiry      time.Duration `mapstructure:"JWT_ACCESS_EXPIRY"`
	RefreshExpiry     time.Duration `mapstructure:"JWT_REFRESH_EXPIRY"`
	DistributorSecret string        `mapstructure:"JWT_DISTRIBUTOR_SECRET"`
}

type OTPConfig struct {
	Length     int           `mapstructure:"OTP_LENGTH"`
	Expiry     time.Duration `mapstructure:"OTP_EXPIRY"`
	MaxRetries int           `mapstructure:"OTP_MAX_RETRIES"`
	// In dev mode, OTPs are logged instead of sent
	DevMode bool `mapstructure:"OTP_DEV_MODE"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"REDIS_ADDR"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"REDIS_DB"`
}

type SurepassConfig struct {
	BaseURL      string `mapstructure:"SUREPASS_BASE_URL"`
	CIBILBaseURL string `mapstructure:"SUREPASS_CIBIL_BASE_URL"`
	Token        string `mapstructure:"SUREPASS_TOKEN"`
	CIBILToken   string `mapstructure:"SUREPASS_CIBIL_TOKEN"`
}

type MSG91Config struct {
	AuthKey    string `mapstructure:"MSG91_AUTH_KEY"`
	TemplateID string `mapstructure:"MSG91_TEMPLATE_ID"`
	SenderID   string `mapstructure:"MSG91_SENDER_ID"`
}

// NotifyConfig controls which notification channels are active.
// Set NOTIFY_EMAIL=true and/or NOTIFY_WHATSAPP=true to enable.
type NotifyConfig struct {
	EmailEnabled    bool   `mapstructure:"NOTIFY_EMAIL"`
	WhatsAppEnabled bool   `mapstructure:"NOTIFY_WHATSAPP"`
	// Email (SMTP)
	SMTPHost     string `mapstructure:"SMTP_HOST"`
	SMTPPort     int    `mapstructure:"SMTP_PORT"`
	SMTPUser     string `mapstructure:"SMTP_USER"`
	SMTPPassword string `mapstructure:"SMTP_PASSWORD"`
	SMTPFrom     string `mapstructure:"SMTP_FROM"`
	// WhatsApp (Meta Cloud API)
	WAAPIURL   string `mapstructure:"WA_API_URL"`
	WAToken    string `mapstructure:"WA_TOKEN"`
	WAPhoneID  string `mapstructure:"WA_PHONE_ID"`
}

type AppConfig struct {
	Env            string `mapstructure:"APP_ENV"` // development | staging | production
	Name           string `mapstructure:"APP_NAME"`
	DocumentBucket string `mapstructure:"DOCUMENT_BUCKET"` // for uploaded screenshots/docs
	BaseURL        string `mapstructure:"BASE_URL"`
}

// Load reads config from environment variables.
// Call this once at startup.
func Load() (*Config, error) {
	v := viper.New()

	// Defaults
	v.SetDefault("PORT", "8081")
	v.SetDefault("READ_TIMEOUT", "60s")
	v.SetDefault("WRITE_TIMEOUT", "60s")
	v.SetDefault("IDLE_TIMEOUT", "120s")
	v.SetDefault("DB_MAX_OPEN_CONNS", 25)
	v.SetDefault("DB_MAX_IDLE_CONNS", 5)
	v.SetDefault("MIGRATIONS_DIR", "./internal/database/migrations")
	v.SetDefault("JWT_ACCESS_EXPIRY", "24h")
	v.SetDefault("JWT_REFRESH_EXPIRY", "168h") // 7 days
	v.SetDefault("OTP_LENGTH", 6)
	v.SetDefault("OTP_EXPIRY", "5m")
	v.SetDefault("OTP_MAX_RETRIES", 5)
	v.SetDefault("OTP_DEV_MODE", false)
	v.SetDefault("REDIS_ADDR", "localhost:6379")
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("APP_NAME", "Kresconet Distributor Credit Platform")
	v.SetDefault("NOTIFY_EMAIL", false)
	v.SetDefault("NOTIFY_WHATSAPP", false)
	v.SetDefault("CORS_ORIGINS", "http://localhost:3000,http://localhost:5173")
	v.SetDefault("SUREPASS_BASE_URL", "https://kyc-api.surepass.io/api/v1")
	v.SetDefault("SUREPASS_CIBIL_BASE_URL", "https://app.surepass.app/production/api/v1")
	v.SetDefault("SHIPROCKET_API_URL", "https://apiv2.shiprocket.in/v1")
	v.SetDefault("SHIPROCKET_PICKUP_LOCATION", "warehouse")

	// Load .env file if present (non-fatal if missing)
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	_ = v.ReadInConfig()

	// Environment variables override file
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	cfg := &Config{}
	cfg.Server.Port = v.GetString("PORT")
	cfg.Server.ReadTimeout = v.GetDuration("READ_TIMEOUT")
	cfg.Server.WriteTimeout = v.GetDuration("WRITE_TIMEOUT")
	cfg.Server.IdleTimeout = v.GetDuration("IDLE_TIMEOUT")
	cfg.Server.CORSOrigins = strings.Split(v.GetString("CORS_ORIGINS"), ",")

	cfg.Database.DSN = v.GetString("DATABASE_URL")
	cfg.Database.MaxOpenConns = v.GetInt("DB_MAX_OPEN_CONNS")
	cfg.Database.MaxIdleConns = v.GetInt("DB_MAX_IDLE_CONNS")
	cfg.Database.ConnMaxLifetime = time.Hour
	cfg.Database.MigrationsDir = v.GetString("MIGRATIONS_DIR")

	cfg.JWT.AccessSecret = v.GetString("JWT_ACCESS_SECRET")
	cfg.JWT.RefreshSecret = v.GetString("JWT_REFRESH_SECRET")
	cfg.JWT.AccessExpiry = v.GetDuration("JWT_ACCESS_EXPIRY")
	cfg.JWT.RefreshExpiry = v.GetDuration("JWT_REFRESH_EXPIRY")
	cfg.JWT.DistributorSecret = v.GetString("JWT_DISTRIBUTOR_SECRET")

	cfg.OTP.Length = v.GetInt("OTP_LENGTH")
	cfg.OTP.Expiry = v.GetDuration("OTP_EXPIRY")
	cfg.OTP.MaxRetries = v.GetInt("OTP_MAX_RETRIES")
	cfg.OTP.DevMode = v.GetBool("OTP_DEV_MODE")

	cfg.Redis.Addr = v.GetString("REDIS_ADDR")
	cfg.Redis.Password = v.GetString("REDIS_PASSWORD")
	cfg.Redis.DB = v.GetInt("REDIS_DB")

	cfg.Surepass.BaseURL = v.GetString("SUREPASS_BASE_URL")
	cfg.Surepass.CIBILBaseURL = v.GetString("SUREPASS_CIBIL_BASE_URL")
	cfg.Surepass.Token = v.GetString("SUREPASS_TOKEN")

	cfg.MSG91.AuthKey = v.GetString("MSG91_AUTH_KEY")
	cfg.MSG91.TemplateID = v.GetString("MSG91_TEMPLATE_ID")
	cfg.MSG91.SenderID = v.GetString("MSG91_SENDER_ID")

	cfg.Notify.EmailEnabled = v.GetBool("NOTIFY_EMAIL")
	cfg.Notify.WhatsAppEnabled = v.GetBool("NOTIFY_WHATSAPP")
	cfg.Notify.SMTPHost = v.GetString("SMTP_HOST")
	cfg.Notify.SMTPPort = v.GetInt("SMTP_PORT")
	cfg.Notify.SMTPUser = v.GetString("SMTP_USER")
	cfg.Notify.SMTPPassword = v.GetString("SMTP_PASSWORD")
	cfg.Notify.SMTPFrom = v.GetString("SMTP_FROM")
	cfg.Notify.WAAPIURL = v.GetString("WA_API_URL")
	cfg.Notify.WAToken = v.GetString("WA_TOKEN")
	cfg.Notify.WAPhoneID = v.GetString("WA_PHONE_ID")

	cfg.Razorpay.KeyID = v.GetString("RAZORPAY_KEY_ID")
	cfg.Razorpay.KeySecret = v.GetString("RAZORPAY_KEY_SECRET")
	cfg.Razorpay.WebhookSecret = v.GetString("RAZORPAY_WEBHOOK_SECRET")

	cfg.Shiprocket.Email = v.GetString("SHIPROCKET_EMAIL")
	cfg.Shiprocket.Password = v.GetString("SHIPROCKET_PASSWORD")
	cfg.Shiprocket.APIToken = v.GetString("SHIPROCKET_API_TOKEN")
	cfg.Shiprocket.APIURL = v.GetString("SHIPROCKET_API_URL")
	cfg.Shiprocket.ChannelID = v.GetString("SHIPROCKET_CHANNEL_ID")
	cfg.Shiprocket.PickupLocation = v.GetString("SHIPROCKET_PICKUP_LOCATION")

	cfg.Admin.Email = v.GetString("ADMIN_EMAIL")
	cfg.Admin.Password = v.GetString("ADMIN_PASSWORD")

	cfg.App.Env = v.GetString("APP_ENV")
	cfg.App.Name = v.GetString("APP_NAME")
	cfg.App.DocumentBucket = v.GetString("DOCUMENT_BUCKET")
	cfg.App.BaseURL = v.GetString("BASE_URL")

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Database.DSN == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.JWT.AccessSecret == "" {
		return fmt.Errorf("JWT_ACCESS_SECRET is required")
	}
	if c.JWT.RefreshSecret == "" {
		return fmt.Errorf("JWT_REFRESH_SECRET is required")
	}
	if c.JWT.DistributorSecret == "" {
		return fmt.Errorf("JWT_DISTRIBUTOR_SECRET is required")
	}
	if c.Admin.Email == "" {
		return fmt.Errorf("ADMIN_EMAIL is required in environment configuration")
	}
	if c.Admin.Password == "" {
		return fmt.Errorf("ADMIN_PASSWORD is required in environment configuration")
	}
	return nil
}

// IsDevelopment returns true when running locally.
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

// IsProduction returns true when running in production.
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}
