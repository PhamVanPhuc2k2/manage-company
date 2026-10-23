package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppName         string        `mapstructure:"APP_NAME"`
	Env             string        `mapstructure:"ENV"`
	LogLeval        string        `mapstructure:"LOG_LEVEL"`
	HTTPPort        int           `mapstructure:"HTTP_PORT"`
	ShutdownTimeout time.Duration `mapstructure:"SHUTDOWN_TIMEOUT"`

	PostgresDSN     string `mapstructure:"POSTGRES_DSN"`
	PostgresMaxConn int32  `mapstructure:"POSTGRES_MAX_CONN"`

	RedisAddr     string `mapstructure:"REDIS_ADDR"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDB       int    `mapstructure:"REDIS_DB"`

	RabbitMQURL       string `mapstructure:"RABBITMQ_URL"`
	WorkerConcurrency int    `mapstructure:"WORKER_CONCURRENCY"`

	JWTSecret string `mapstructure:"JWT_SECRET"`

	IntegrationSecretKey string        `mapstructure:"INTEGRATION_SECRET_KEY"`
	JWTAccessTTL         time.Duration `mapstructure:"JWT_ACCESS_TTL"`
	JWTRefreshTTL        time.Duration `mapstructure:"JWT_REFRESH_TTL"`

	AuthOTPEnabled bool `mapstructure:"AUTH_OTP_ENABLED"`

	CORSAllowedOrigins []string `mapstructure:"CORS_ALLOWED_ORIGINS"`

	PublicBaseURL string `mapstructure:"PUBLIC_BASE_URL"`

	SMTPHost string `mapstructure:"SMTP_HOST"`
	SMTPPort int    `mapstructure:"SMTP_PORT"`
	SMTPFrom string `mapstructure:"SMTP_FROM"`
	SMTPUser string `mapstructure:"SMTP_USER"`
	SMTPPass string `mapstructure:"SMTP_PASS"`

	R2AccountID       string `mapstructure:"R2_ACCOUNT_ID"`
	R2AccessKeyID     string `mapstructure:"R2_ACCESS_KEY_ID"`
	R2SecretAccessKey string `mapstructure:"R2_SECRET_ACCESS_KEY"`
	R2Bucket          string `mapstructure:"R2_BUCKET"`
	R2PublicURL       string `mapstructure:"R2_PUBLIC_URL"`

	VAPIDPublicKey  string `mapstructure:"VAPID_PUBLIC_KEY"`
	VAPIDPrivateKey string `mapstructure:"VAPID_PRIVATE_KEY"`
	VAPIDSubject    string `mapstructure:"VAPID_SUBJECT"`
}

func (c Config) IsProduction() bool { return c.Env == "production" }

func Load(appName string) (*Config, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	setDefault := func(key string, value any) {
		v.SetDefault(key, value)
		_ = v.BindEnv(key)
	}

	setDefault("APP_NAME", appName)
	setDefault("ENV", "development")
	setDefault("LOG_LEVEL", "debug")
	setDefault("HTTP_PORT", 8080)
	setDefault("SHUTDOWN_TIMEOUT", "30s")

	setDefault("POSTGRES_DSN", "")
	setDefault("POSTGRES_MAX_CONN", 20)

	setDefault("REDIS_ADDR", "redis:6379")
	setDefault("REDIS_PASSWORD", "")
	setDefault("REDIS_DB", 0)

	setDefault("RABBITMQ_URL", "")
	setDefault("WORKER_CONCURRENCY", 5)

	setDefault("JWT_SECRET", "")
	setDefault("JWT_ACCESS_TTL", "15m")
	setDefault("JWT_REFRESH_TTL", "168h")
	setDefault("AUTH_OTP_ENABLED", true)
	setDefault("INTEGRATION_SECRET_KEY", "")

	setDefault("CORS_ALLOWED_ORIGINS", []string{"http://localhost", "http://localhost:3000"})
	setDefault("PUBLIC_BASE_URL", "http://localhost")

	setDefault("SMTP_HOST", "mailhog")
	setDefault("SMTP_PORt", 1025)
	setDefault("SMTP_FROM", "no-reply@manage-company.local")
	setDefault("SMTP_USER", "")
	setDefault("SMTP_PASS", "")

	setDefault("R2_ACCOUNT_ID", "")
	setDefault("R2_ACCESS_KEY_ID", "")
	setDefault("R2_SECRET_ACCESS_KEY", "")
	setDefault("R2_BUCKET", "")
	setDefault("R2_PUBLIC_URl", "")

	setDefault("VAPID_PUBLIC_KEY", "")
	setDefault("VAPID_PRIVATE_KEY", "")
	setDefault("VAPID_SUBJECT", "")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("giải mã cấu hình: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c Config) validate() error {
	var missing []string

	if c.PostgresDSN == "" {
		missing = append(missing, "POSTGRES_DSN")
	}
	if c.RabbitMQURL == "" {
		missing = append(missing, "RABBITMQ_URL")
	}
	if c.IsProduction() {
		if c.JWTSecret == "" || strings.HasPrefix(c.JWTSecret, "doi_chuoi_nay") {
			missing = append(missing, "JWT_SECRET (không đươc dùng giá trị mặc định ở production)")
		}
		if c.IntegrationSecretKey == "" || strings.HasPrefix(c.IntegrationSecretKey, "doi_chuoi_nay") {
			missing = append(missing, "INTEGRATION_SECRET_KEY (không được dùng giá trị mặc định ở production)")
		}
		if c.AuthOTPEnabled && (c.SMTPHost == "" || c.SMTPHost == "mailhog") {
			missing = append(missing, "SMTP_HOST (máy chủ gửi email thật - cần để gửi mã OTP đăng nhập)")
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("thiếu cấu hình bắt buộc: %s", strings.Join(missing, ", "))
	}

	return nil
}

func (c Config) VAPIDContact() string {
	if c.VAPIDSubject != "" {
		return c.VAPIDSubject
	}
	return "mailto" + c.JWTSecret
}
