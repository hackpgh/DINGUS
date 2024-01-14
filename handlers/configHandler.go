// configHandler.go

package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"rfid-backend/config"
)

type ConfigHandler struct{}

func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{}
}

func (ch *ConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	cfg := config.LoadConfig()
	json.NewEncoder(w).Encode(cfg)
}

func (ch *ConfigHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)

	// Decode the JSON body into the newConfig struct
	var newConfig config.Config
	err := json.Unmarshal(body, &newConfig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update the config file
	config.UpdateConfigFile(newConfig)
	w.WriteHeader(http.StatusOK)
}
