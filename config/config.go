package config

import (
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

/*
|--------------------------------------------------------------------------
| Config Structs
|--------------------------------------------------------------------------
*/

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	App      App            `mapstructure:"app"`
	OTPCode  OTPCode        `mapstructure:"OTPCode"`
}

type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	Mode         string        `mapstructure:"mode"` // debug, release, test
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Name            string        `mapstructure:"name"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	LogLevel        string        `mapstructure:"log_level"` // silent, error, warn, info
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	Database int    `mapstructure:"database"`
	PoolSize int    `mapstructure:"pool_size"`
}

type JWTConfig struct {
	AccessTokenExpireDuration  time.Duration `mapstructure:"access_token_duration"`
	RefreshTokenExpireDuration time.Duration `mapstructure:"refresh_token_duration"`
	Secret                     string        `mapstructure:"secret"`
	RefreshSecret              string        `mapstructure:"refresh_secret"`
}

type App struct {
	Environment string `mapstructure:"environment"` // development, production, test
	LogLevel    string `mapstructure:"log_level"`   // debug, info, warn, error
}

type OTPCode struct {
	ExpireTime time.Duration `mapstructure:"expire_time"`
	TryAttempt int           `mapstructure:"try_attempt"`
}

var Cfg *Config

func Load() {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		log.Println("⚠️ config file not found, using environment variables only")
	}

	Cfg = &Config{}
	if err := viper.Unmarshal(Cfg); err != nil {
		log.Fatalf("❌ failed to unmarshal config: %v", err)
	}

	logConfig()
}

func setDefaults() {
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.read_timeout", 10*time.Second)
	viper.SetDefault("server.write_timeout", 10*time.Second)

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.password", "app_user_password")
	viper.SetDefault("database.user", "app_user")
	viper.SetDefault("database.port", "3306")
	viper.SetDefault("database.name", "app_db")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.conn_max_lifetime", 60*time.Minute)

	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.database", 0)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.password", "StrongPasswordHere")

	viper.SetDefault("jwt.access_token_duration", 15*time.Minute)
	viper.SetDefault("jwt.refresh_token_duration", 24*time.Hour)

	viper.SetDefault("app.environment", "debug")
	viper.SetDefault("app.log_level", "debug")
	viper.SetDefault("expire_time", 3*time.Minute)
	viper.SetDefault("try_attempt", 3)
}

/*
|--------------------------------------------------------------------------
| Logger
|--------------------------------------------------------------------------
*/

func logConfig() {
	if Cfg.App.Environment != "development" {
		return
	}

	log.Println("📦 Config Loaded")
	log.Printf("  ENV: %s", Cfg.App.Environment)
	log.Printf("  Server: :%s (%s)", Cfg.Server.Port, Cfg.Server.Mode)
	log.Printf("  DB: %s:%s/%s", Cfg.Database.Host, Cfg.Database.Port, Cfg.Database.Name)
	log.Printf("  Redis: %s:%s", Cfg.Redis.Host, Cfg.Redis.Port)
}
