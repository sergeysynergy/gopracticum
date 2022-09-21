package config

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/sergeysynergy/metricser/pkg/crypter"
	"log"
	"os"
	"strings"
	"time"
)

type Duration struct {
	time.Duration
}

func (d *Duration) MarshalJSON() ([]byte, error) {
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

type ServerConf struct {
	Addr            string        `env:"ADDRESS" json:"address"`
	GRPCAddr        string        `env:"GRPC_ADDRESS" json:"grpc_addr"`
	StoreFile       string        `env:"STORE_FILE" json:"store_file"`
	Restore         bool          `env:"RESTORE" json:"restore"`
	MyStoreInterval Duration      `json:"store_interval"`
	StoreInterval   time.Duration `env:"STORE_INTERVAL"`
	DatabaseDSN     string        `env:"DATABASE_DSN" json:"database_dsn"`
	CryptoKey       string        `env:"CRYPTO_KEY" json:"crypto_key"`
	Key             string        `env:"KEY"`
	TrustedSubnet   string        `env:"TRUSTED_SUBNET"`
	ConfigFile      string
	PrivateKey      *rsa.PrivateKey
}

func NewServerConf() *ServerConf {
	defaultCfg := &ServerConf{
		Addr:          "127.0.0.1:8080",
		GRPCAddr:      ":3200",
		StoreFile:     "/tmp/devops-metrics-pgsql.json",
		Restore:       true,
		StoreInterval: 300 * time.Second,
	}

	if cfgFile, ok := getConfigFile(); ok {
		cfg := &ServerConf{}
		err := LoadFromFile(cfgFile, cfg)
		if err != nil {
			log.Println("[ERROR]", err)
		} else {
			log.Println("[DEBUG] Using config file:", cfgFile)
			return cfg
		}
	}

	return defaultCfg
}

func (c *ServerConf) Init() {
	pk, err := crypter.OpenPrivate(c.CryptoKey)
	if err != nil {
		log.Println("[WARNING] Failed to get private key -", err)
	} else {
		c.PrivateKey = pk
		log.Println("[DEBUG] Encryption enabled using RSA")
	}
}

type AgentConfig struct {
	Addr             string        `env:"ADDRESS" json:"address"`
	GRPCAddr         string        `env:"GRPC_ADDRESS" json:"grpc_addr"`
	MyReportInterval Duration      `json:"report_interval"`
	MyPollInterval   Duration      `json:"poll_interval"`
	ReportInterval   time.Duration `env:"REPORT_INTERVAL"`
	PollInterval     time.Duration `env:"POLL_INTERVAL"`
	Key              string        `env:"KEY"`
	CryptoKey        string        `env:"CRYPTO_KEY"`
	ConfigFile       string
}

func NewAgentConf() *AgentConfig {
	defaultCfg := &AgentConfig{
		Addr:           "127.0.0.1:8080",
		GRPCAddr:       ":3200",
		ReportInterval: 10 * time.Second,
		PollInterval:   2 * time.Second,
	}

	if cfgFile, ok := getConfigFile(); ok {
		cfg := &AgentConfig{}
		err := LoadFromFile(cfgFile, cfg)
		if err != nil {
			log.Println("[ERROR]", err)
		} else {
			log.Println("[DEBUG] Using config file:", cfgFile)
			return cfg
		}
	}

	return defaultCfg
}

// Получим путь к файлу из аргументов или переменной окружения.
func getConfigFile() (string, bool) {
	cfgFile, ok := os.LookupEnv("CONFIG")
	if !ok {
		for k, v := range os.Args[1:] {
			if v == "-c" && len(os.Args) > k+2 {
				cfgFile = os.Args[k+2]
			}
			if v == "-config" && len(os.Args) > k+2 {
				cfgFile = os.Args[k+2]
			}
			if strings.HasPrefix(v, "-c=") {
				cfgFile = os.Args[k+1][3:]
			}
			if strings.HasPrefix(v, "-config=") {
				cfgFile = os.Args[k+1][8:]
			}
		}
	}

	if cfgFile == "" {
		return cfgFile, false
	}

	return cfgFile, true
}

// LoadFromFile Получаем значения конфига из JSON-файла.
func LoadFromFile(filePath string, cfg interface{}) (err error) {
	if filePath == "" {
		return fmt.Errorf("empty file name")
	}
	f, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}

	err = json.Unmarshal(f, &cfg)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	return nil
}
