package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	CertFile                 string `yaml:"cert_file"`
	ContactRFIDCacheFilePath string `yaml:"contact_rfid_cache_file_path"`
	DatabasePath             string `yaml:"database_path"`
	KeyFile                  string `yaml:"key_file"`
	WildApricotAccountId     int    `yaml:"wild_apricot_account_id"`
	RFIDFieldName            string `yaml:"rfid_field_name"`
	TrainingFieldName        string `yaml:"training_field_name"`
}

func LoadConfig() *Config {
	viper.SetConfigName("config") // Name of the config file (without extension)
	viper.SetConfigType("yml")    // Config file type
	viper.AddConfigPath(".")      // Path to look for the config file in

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Unable to decode into struct: %s", err)
	}

	return &cfg
}
