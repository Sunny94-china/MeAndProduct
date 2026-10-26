package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv    string `mapstructure:"APP_ENV"`
	Server    ServerCfg
	MySQL     MySQLCfg
	Redis     RedisCfg
	JWT       JWTCfg
	CORS      CORSCfg
	Migration MigrationCfg
	RateLimit RateLimitCfg
	Admin     AdminCfg
}

type ServerCfg struct {
	Port         string        `mapstructure:"PORT"`
	ReadTimeout  time.Duration `mapstructure:"READ_TIMEOUT"`
	WriteTimeout time.Duration `mapstructure:"WRITE_TIMEOUT"`
}

type MySQLCfg struct {
	Host         string `mapstructure:"HOST"`
	Port         int    `mapstructure:"PORT"`
	User         string `mapstructure:"USER"`
	Password     string `mapstructure:"PASSWORD"`
	DB           string `mapstructure:"DB"`
	MaxOpenConns int    `mapstructure:"MAX_OPEN_CONNS"`
	MaxIdleConns int    `mapstructure:"MAX_IDLE_CONNS"`
	MigrateAuto  bool   `mapstructure:"MIGRATE_AUTO"`
}

func (m MySQLCfg) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local&multiStatements=true",
		m.User, m.Password, m.Host, m.Port, m.DB,
	)
}

type RedisCfg struct {
	Addr     string `mapstructure:"ADDR"`
	Password string `mapstructure:"PASSWORD"`
	DB       int    `mapstructure:"DB"`
	PoolSize int    `mapstructure:"POOL_SIZE"`
}

type JWTCfg struct {
	Secret             string        `mapstructure:"SECRET"`
	AccessTTL          time.Duration `mapstructure:"ACCESS_TTL"`
	RefreshTTL         time.Duration `mapstructure:"REFRESH_TTL"`
	Issuer             string        `mapstructure:"ISSUER"`
	BlacklistKeyPrefix string        `mapstructure:"BLACKLIST_PREFIX"`
}

type CORSCfg struct {
	AllowOrigins     string `mapstructure:"ALLOW_ORIGINS"` // comma separated
	AllowMethods     string `mapstructure:"ALLOW_METHODS"`
	AllowHeaders     string `mapstructure:"ALLOW_HEADERS"`
	AllowCredentials bool   `mapstructure:"ALLOW_CREDENTIALS"`
	MaxAge           int    `mapstructure:"MAX_AGE"`
}

func (c CORSCfg) OriginsSlice() []string {
	return splitTrim(c.AllowOrigins)
}
func (c CORSCfg) MethodsSlice() []string { return splitTrim(c.AllowMethods) }
func (c CORSCfg) HeadersSlice() []string { return splitTrim(c.AllowHeaders) }

type MigrationCfg struct {
	Path string `mapstructure:"PATH"`
}

type AdminCfg struct {
	Username string `mapstructure:"USERNAME"`
	Email    string `mapstructure:"EMAIL"`
	Password string `mapstructure:"PASSWORD"`
}

type RateLimitCfg struct {
	LoginLimit    int           `mapstructure:"LOGIN_LIMIT"`
	LoginWindow   time.Duration `mapstructure:"LOGIN_WINDOW"`
	ReportLimit   int           `mapstructure:"REPORT_LIMIT"`
	ReportWindow  time.Duration `mapstructure:"REPORT_WINDOW"`
	GeneralLimit  int           `mapstructure:"GENERAL_LIMIT"`
	GeneralWindow time.Duration `mapstructure:"GENERAL_WINDOW"`
}

func splitTrim(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// Load reads config from (optional) yaml file and env vars.
// Env vars take precedence. Nested keys use underscore: MYSQL_HOST=...
func Load(configFile ...string) (*Config, error) {
	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	if len(configFile) > 0 && configFile[0] != "" {
		v.SetConfigFile(configFile[0])
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("read config file: %w", err)
		}
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("SERVER.PORT", "8080")
	v.SetDefault("SERVER.READ_TIMEOUT", "30s")
	v.SetDefault("SERVER.WRITE_TIMEOUT", "30s")

	v.SetDefault("MYSQL.HOST", "127.0.0.1")
	v.SetDefault("MYSQL.PORT", 3306)
	v.SetDefault("MYSQL.USER", "meandproduct")
	v.SetDefault("MYSQL.PASSWORD", "dev_pwd")
	v.SetDefault("MYSQL.DB", "meandproduct")
	v.SetDefault("MYSQL.MAX_OPEN_CONNS", 50)
	v.SetDefault("MYSQL.MAX_IDLE_CONNS", 10)
	v.SetDefault("MYSQL.MIGRATE_AUTO", true)

	v.SetDefault("REDIS.ADDR", "127.0.0.1:6379")
	v.SetDefault("REDIS.PASSWORD", "dev_redis_pwd")
	v.SetDefault("REDIS.DB", 0)
	v.SetDefault("REDIS.POOL_SIZE", 20)

	v.SetDefault("JWT.SECRET", "change-me-in-production-please-32chars-min")
	v.SetDefault("JWT.ACCESS_TTL", "2h")
	v.SetDefault("JWT.REFRESH_TTL", "168h") // 7 days
	v.SetDefault("JWT.ISSUER", "meandproduct")
	v.SetDefault("JWT.BLACKLIST_PREFIX", "mp:jwt:blacklist:")

	v.SetDefault("CORS.ALLOW_ORIGINS", "*")
	v.SetDefault("CORS.ALLOW_METHODS", "GET,POST,PUT,DELETE,OPTIONS,PATCH")
	v.SetDefault("CORS.ALLOW_HEADERS", "Origin,Content-Type,Accept,Authorization,X-Request-Id")
	v.SetDefault("CORS.ALLOW_CREDENTIALS", true)
	v.SetDefault("CORS.MAX_AGE", 86400)

	v.SetDefault("MIGRATION.PATH", "file://migrations")

	v.SetDefault("RATELIMIT.LOGIN_LIMIT", 60)
	v.SetDefault("RATELIMIT.LOGIN_WINDOW", "5m")
	v.SetDefault("RATELIMIT.REPORT_LIMIT", 5)
	v.SetDefault("RATELIMIT.REPORT_WINDOW", "1h")
	v.SetDefault("RATELIMIT.GENERAL_LIMIT", 300)
	v.SetDefault("RATELIMIT.GENERAL_WINDOW", "1m")

	v.SetDefault("ADMIN.USERNAME", "admin")
	v.SetDefault("ADMIN.EMAIL", "admin@local")
	v.SetDefault("ADMIN.PASSWORD", "Admin@123")
}
