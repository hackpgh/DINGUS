// configHandler.go

package handlers

import (
	"net/http"
	"rfid-backend/config"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct{}

func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{}
}

// func (ch *ConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
// 	cfg := config.LoadConfig()
// 	json.NewEncoder(w).Encode(cfg)
// }

func (ch *ConfigHandler) UpdateConfig(c *gin.Context) {
	var newConfig config.Config

	if err := c.BindJSON(&newConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := config.UpdateConfigFile(newConfig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update configuration"})
		return
	}

	c.Status(http.StatusOK)
}
