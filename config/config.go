package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	CertFile             string `yaml:"cert_file"`
	KeyFile              string `yaml:"key_file"`
	DatabasePath         string `yaml:"database_path"`
	WildApricotAccountId int    `yaml:"wild_apricot_account_id"`
	RFIDFieldName        string `yaml:"rfid_field_name"`
	TrainingFieldName    string `yaml:"training_field_name"`
}

func LoadConfig() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	//TODO: Hacky. Fix the viper unmarshalling bug
	configMap := viper.AllSettings()
	cfg := Config{
		CertFile:             configMap["cert_file"].(string),
		KeyFile:              configMap["key_file"].(string),
		DatabasePath:         configMap["database_path"].(string),
		WildApricotAccountId: configMap["wild_apricot_account_id"].(int),
		RFIDFieldName:        configMap["rfid_field_name"].(string),
		TrainingFieldName:    configMap["training_field_name"].(string),
	}

	return &cfg
}
