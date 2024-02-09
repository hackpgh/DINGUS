// file: /pkg/config/config.go
package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"dingus/pkg/utils"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	config *utils.Singleton
	once   sync.Once
)

type Config struct {
	// TLS
	CertFile   string `mapstructure:"cert_file"`
	KeyFile    string `mapstructure:"key_file"`
	UseAutoTLS bool
	// Database
	DatabasePath string
	// Wild Apricot
	TagIdFieldName          string `mapstructure:"tag_id_field_name"`
	TrainingFieldName       string `mapstructure:"training_field_name"`
	WildApricotAccountId    int    `mapstructure:"wild_apricot_account_id"`
	ContactFilterQuery      string `mapstructure:"contact_filter_query"`
	WildApricotApiKey       string `mapstructure:"WILD_APRICOT_API_KEY"`
	WildApricotWebhookToken string
	// SSO
	SSOClientID       string `mapstructure:"sso_client_id"`
	SSOClientSecret   string `mapstructure:"sso_client_secret"`
	SSORedirectURI    string `mapstructure:"sso_redirect_uri"`
	CookieStoreSecret string `mapstructure:"cookie_store_secret"`
	// Logging
	LogLevel    string `mapstructure:"log_level"`
	LokiHookURL string `mapstructure:"loki_hook_url"`
	// Logger instance
	log *logrus.Logger
}

func init() {
	config = utils.NewSingleton(loadConfig)
}

func LoadConfig() *Config {
	return config.Get(loadConfig).(*Config)
}

func loadConfig() interface{} {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found")
	}
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// Read configuration from config file first
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file, %s", err)
	}

	viper.AutomaticEnv()         // Read configuration from environment variables
	fmt.Println(viper.AllKeys()) // This will print all keys viper knows about

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Error unmarshalling config: %s", err)
	}

	fmt.Printf("Config loaded: %+v\n", cfg)

	// After loading from both config file and environment, compose the DSN
	cfg.DatabasePath = getPostgresConnectionString()

	// Resolve paths for any files (e.g., TLS certificates)
	cfg.resolvePaths()

	return &cfg
}

func (cfg *Config) resolvePaths() {
	var err error
	fmt.Printf("cert: %s", cfg.CertFile)
	cfg.CertFile, err = filepath.Abs(cfg.CertFile)
	fmt.Printf("cert: %s", cfg.CertFile)
	if err != nil || !fileExists(cfg.CertFile) {
		log.Fatalf("Certificate file not found or path is invalid: %s", cfg.CertFile)
	}

	cfg.KeyFile, err = filepath.Abs(cfg.KeyFile)
	if err != nil || !fileExists(cfg.KeyFile) {
		log.Fatalf("Key file not found or path is invalid: %s", cfg.KeyFile)
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := os.Getenv(name)
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultVal
}

func UpdateConfigFile(newConfig Config) error {
	projectRoot, err := utils.GetProjectRoot()
	if err != nil {
		log.Fatalf("Error fetching project root absolute path: %s", err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(projectRoot)

	log.Printf("Attempting to read existing config for update")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	if newConfig.CertFile != "" {
		viper.Set("cert_file", newConfig.CertFile)
	}
	if newConfig.KeyFile != "" {
		viper.Set("key_file", newConfig.KeyFile)
	}

	if newConfig.WildApricotAccountId != 0 {
		viper.Set("wild_apricot_account_id", newConfig.WildApricotAccountId)

	}
	if newConfig.WildApricotAccountId != 0 {
		viper.Set("contact_filter_query", newConfig.ContactFilterQuery)

	}
	if newConfig.TagIdFieldName != "" {
		viper.Set("tag_id_field_name", newConfig.TagIdFieldName)
	}
	if newConfig.TrainingFieldName != "" {
		viper.Set("training_field_name", newConfig.TrainingFieldName)
	}

	err = viper.WriteConfig()
	if err != nil {
		log.Fatalf("Error writing to config file: %s", err)
		return err
	}
	log.Println("Configuration file updated successfully.")
	return nil
}

func getPostgresConnectionString() string {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")
	sslmode := os.Getenv("POSTGRES_SSLMODE")

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
}
