package config

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"rfid-backend/utils"

	"github.com/spf13/viper"
)

var (
	config *utils.Singleton
	once   sync.Once
)

type Config struct {
	CertFile             string `mapstructure:"cert_file" json:"cert_file"`
	DatabasePath         string `mapstructure:"database_path" json:"database_path"`
	KeyFile              string `mapstructure:"key_file" json:"key_file"`
	RFIDFieldName        string `mapstructure:"rfid_field_name" json:"rfid_field_name"`
	TrainingFieldName    string `mapstructure:"training_field_name" json:"training_field_name"`
	WildApricotAccountId int    `mapstructure:"wild_apricot_account_id" json:"wild_apricot_account_id"`
	ContactFilterQuery   string `mapstructure:"contact_filter_query" json:"contact_filter_query"`
}

func init() {
	// Initialize the config singleton instance.
	config = utils.NewSingleton(loadConfig())
}

// LoadConfig returns the configuration instance.
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

	return &cfg
}

// UpdateConfigFile updates the configuration settings based on the provided newConfig.
func UpdateConfigFile(newConfig Config) {
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

	// Update viper's settings only if the newConfig fields are not empty
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
	if newConfig.RFIDFieldName != "" {
		viper.Set("rfid_field_name", newConfig.RFIDFieldName)
	}
	if newConfig.TrainingFieldName != "" {
		viper.Set("training_field_name", newConfig.TrainingFieldName)
	}

	// Save the new settings back to the config file
	err = viper.WriteConfig()
	if err != nil {
		log.Fatalf("Error writing to config file: %s", err)
	} else {
		log.Println("Configuration file updated successfully.")
	}
}
