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
	App      App            `mapstructure:"app"`
	Rabbit   RabbitConfig   `mapstructure:"rabbit"`
	GRPC     GRPCConfig     `mapstructure:"grpc"`
	Storage  MinioConfig    `mapstructure:"storage"`
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

type RabbitConfig struct {
	Host            string `mapstructure:"host"`
	Port            string `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	WorkerCount     int    `mapstructure:"worker_count"`
	MaxFFmpegWorker int    `mapstructure:"max_ffmpeg_worker"`
	JobTimeout      int    `mapstructure:"job_timeout_minutes"`
}
type GRPCConfig struct {
	Host string `mapstructure:"grpc_host"`
	Port string `mapstructure:"grpc_port"`
}

type MinioConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl"`
}

type App struct {
	Environment string `mapstructure:"environment"` // debug, release,  test
	LogLevel    string `mapstructure:"log_level"`   // debug, info, warn, error
}

var Cfg *Config

func Load() {
	viper.SetConfigFile("../.env")
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

	viper.SetDefault("app.environment", "debug")
	viper.SetDefault("app.log_level", "debug")

	viper.SetDefault("rabbit.host", "localhost")
	viper.SetDefault("rabbit.port", "5672")
	viper.SetDefault("rabbit.user", "guest")
	viper.SetDefault("rabbit.password", "guest")
	viper.SetDefault("rabbit.worker_count", 6)
	viper.SetDefault("rabbit.max_ffmpeg_worker", 4)
	viper.SetDefault("rabbit.job_timeout_minutes", 30)

	viper.SetDefault("grpc.grpc_host", "localhost")
	viper.SetDefault("grpc.grpc_port", 50051)

	viper.SetDefault("storage.endpoint", "localhost:9000")
	viper.SetDefault("storage.access_key", "admin")
	viper.SetDefault("storage.secret_key", "StrongPassword123!")
	viper.SetDefault("storage.bucket", "videos")
	viper.SetDefault("storage.use_ssl", false)
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
