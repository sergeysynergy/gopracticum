package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) (err error) {
	var value string
	if err = json.Unmarshal(b, &value); err != nil {
		return err
	}

	d.Duration, err = time.ParseDuration(value)
	if err != nil {
		return err
	}

	return nil
}

type Config struct {
	Addr          string   `env:"ADDRESS" json:"address"`
	StoreFile     string   `env:"STORE_FILE" json:"store_file"`
	Restore       bool     `env:"RESTORE" json:"restore"`
	StoreInterval Duration `env:"STORE_INTERVAL" json:"store_interval"`
	DatabaseDSN   string   `env:"DATABASE_DSN" json:"database_dsn"`
	CryptoKey     string   `env:"CRYPTO_KEY" json:"crypto_key"`
	Key           string   `env:"KEY"`
	ConfigFile    string
}

func New() *Config {
	cfg := &Config{
		Addr:      "127.0.0.1:8080",
		StoreFile: "/tmp/devops-metrics-pgsql.json",
		Restore:   true,
	}
	cfg.StoreInterval.Duration = 300 * time.Second

	return cfg
}

// LoadFromFile Получаем значения конфига из JSON-файла.
func LoadFromFile(filePath string) (*Config, error) {
	if filePath == "" {
		return nil, fmt.Errorf("empty file name")
	}
	f, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}

	cfg := &Config{}
	err = json.Unmarshal(f, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	return cfg, nil
}
