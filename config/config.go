package config

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"rfid-backend/utils"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	config *utils.Singleton
	once   sync.Once
)

type Config struct {
	// TLS
	CertFile     string `mapstructure:"cert_file" json:"cert_file"`
	DatabasePath string `mapstructure:"database_path" json:"database_path"`
	KeyFile      string `mapstructure:"key_file" json:"key_file"`
	// Wild Apricot
	TagIdFieldName          string `mapstructure:"tag_id_field_name" json:"tag_id_field_name"`
	TrainingFieldName       string `mapstructure:"training_field_name" json:"training_field_name"`
	WildApricotAccountId    int    `mapstructure:"wild_apricot_account_id" json:"wild_apricot_account_id"`
	ContactFilterQuery      string `mapstructure:"contact_filter_query" json:"contact_filter_query"`
	WildApricotApiKey       string
	WildApricotWebhookToken string
	// SSO
	SSOClientID       string `mapstructure:"sso_client_id" json:"sso_client_id"`
	SSOClientSecret   string `mapstructure:"sso_client_secret" json:"sso_client_secret"`
	SSORedirectURI    string `mapstructure:"sso_redirect_uri" json:"sso_redirect_uri"`
	CookieStoreSecret string `mapstructure:"cookie_store_secret" json:"cookie_store_secret"`
	// Logging
	log         *logrus.Logger
	LokiHookURL string `mapstructure:"loki_hook_url" json:"loki_hook_url"`
}

func init() {
	config = utils.NewSingleton(loadConfig())
}

func LoadConfig() *Config {
	return config.Get(loadConfig).(*Config)
}

func loadConfig() interface{} {
	projectRoot, err := utils.GetProjectRoot()
	if err != nil {
		log.Fatalf("Error fetching project root absolute path: %s", err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(projectRoot)

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Error unmarshalling config file: %s", err)
	}

	// Resolve relative paths
	cfg.CertFile = filepath.Join(projectRoot, cfg.CertFile)
	if _, err := os.Stat(cfg.CertFile); os.IsNotExist(err) {
		log.Fatalf("Certificate file not found: %s", cfg.CertFile)
	}

	cfg.KeyFile = filepath.Join(projectRoot, cfg.KeyFile)
	if _, err := os.Stat(cfg.KeyFile); os.IsNotExist(err) {
		log.Fatalf("Key file not found: %s", cfg.KeyFile)
	}

	// Load environment variables
	cfg.WildApricotApiKey = os.Getenv("WILD_APRICOT_API_KEY")
	if cfg.WildApricotApiKey == "" {
		log.Fatalf("WILD_APRICOT_API_KEY not set in environment variables")
	}

	cfg.WildApricotWebhookToken = os.Getenv("WILD_APRICOT_WEBHOOK_TOKEN")
	if cfg.WildApricotWebhookToken == "" {
		log.Fatalf("WILD_APRICOT_WEBHOOK_TOKEN not set in environment variables")
	}

	cfg.SSOClientID = os.Getenv("WILD_APRICOT_SSO_CLIENT_ID")
	if cfg.SSOClientID == "" {
		log.Fatalf("WILD_APRICOT_SSO_CLIENT_ID not set in environment variables")
	}

	cfg.SSOClientSecret = os.Getenv("WILD_APRICOT_SSO_CLIENT_SECRET")
	if cfg.SSOClientSecret == "" {
		log.Fatalf("WILD_APRICOT_SSO_CLIENT_SECRET not set in environment variables")
	}

	// TODO: Move this to the config yaml file?
	cfg.SSORedirectURI = os.Getenv("WILD_APRICOT_SSO_REDIRECT_URI")
	if cfg.SSORedirectURI == "" {
		log.Fatalf("WILD_APRICOT_SSO_REDIRECT_URI not set in environment variables")
	}

	cfg.CookieStoreSecret = os.Getenv("COOKIE_STORE_SECRET")
	if cfg.CookieStoreSecret == "" {
		log.Fatalf("COOKIE_STORE_SECRET not set in environment variables")
	}

	return &cfg
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
	if newConfig.DatabasePath != "" {
		viper.Set("database_path", newConfig.DatabasePath)
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
