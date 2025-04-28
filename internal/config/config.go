package config

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type Config struct {
	OrderExpirationDays int    `yaml:"order_expiration_days"`
	FilterWord          string `yaml:"filter_word"`
	TestCredentials     struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"test_credentials"`

	DBUser            string `yaml:"db_user"`
	DBPass            string `yaml:"db_pass"`
	DBHost            string `yaml:"db_host"`
	DBName            string `yaml:"db_name"`
	DBPort            string `yaml:"db_port"`
	ListenAddress     string `yaml:"listen_address"`
	GRPCListenAddress string `yaml:"grpc_listen_address"`

	Interval     int    `yaml:"cron_job_interval"`
	MemCacheHost string `yaml:"memcache_host"`

	MetricsPort string `yaml:"metrics_port"`

	Kafka struct {
		Brokers             []string `yaml:"brokers"`
		Topic               string   `yaml:"topic"`
		GroupID             string   `yaml:"group_id"`
		Partition           int32    `yaml:"partition"`
		BufferSize          int      `yaml:"buffer_size"`
		ReadTimeoutSeconds  int      `yaml:"read_timeout_seconds"`
		WriteTimeoutSeconds int      `yaml:"write_timeout_seconds"`
	} `yaml:"kafka"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.ListenAddress == "" {
		cfg.ListenAddress = "localhost:9000"
	}

	return &cfg, nil
}

func ApplyEnvironmentVariables(cfg *Config) {
	if memcachedHost := os.Getenv("MEMCACHED_HOST"); memcachedHost != "" {
		log.Printf("MEMCACHED_HOST value was set to: %s", memcachedHost)
		cfg.MemCacheHost = memcachedHost
	}
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		log.Printf("DB_HOST value was set to: %s", dbHost)
		cfg.DBHost = dbHost
	}
}
