package config

import (
	"log"
	"sync"

	"rfid-backend/utils"

	"github.com/spf13/viper"
)

var (
	config *utils.Singleton
	once   sync.Once
)

type Config struct {
	CertFile             string `yaml:"cert_file" json:"cert_file"`
	KeyFile              string `yaml:"key_file" json:"key_file"`
	DatabasePath         string `yaml:"database_path" json:"database_path"`
	WildApricotAccountId int    `yaml:"wild_apricot_account_id" json:"wild_apricot_account_id"`
	ContactFilterQuery   string `yaml:"contact_filter_query" json:"contact_filter_query"`
	RFIDFieldName        string `yaml:"rfid_field_name" json:"rfid_field_name"`
	TrainingFieldName    string `yaml:"training_field_name" json:"training_field_name"`
}

func init() {
	// Initialize the config singleton instance.
	config = utils.NewSingleton(loadConfig())
}

// LoadConfig returns the configuration instance.
func LoadConfig() *Config {
	return config.Get(loadConfig).(*Config)
}

// loadConfig is the internal function used to load the configuration settings.
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

	configMap := viper.AllSettings()
	cfg := &Config{
		CertFile:             configMap["cert_file"].(string),
		KeyFile:              configMap["key_file"].(string),
		DatabasePath:         configMap["database_path"].(string),
		WildApricotAccountId: configMap["wild_apricot_account_id"].(int),
		ContactFilterQuery:   configMap["contact_filter_query"].(string),
		RFIDFieldName:        configMap["rfid_field_name"].(string),
		TrainingFieldName:    configMap["training_field_name"].(string),
	}

	return cfg
}

// UpdateConfigFile updates the configuration settings based on the provided newConfig.
func UpdateConfigFile(newConfig Config) {
	projectRoot, err := utils.GetProjectRoot()
	if err != nil {
		log.Fatalf("Error fetching project root absolute path: %s", err)
	}

	// Reload the config file to refresh Viper's internal state
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(projectRoot)

	log.Printf("newConfig: %v", newConfig)
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
