package handlers

import (
	"net/http"
	"rfid-backend/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ConfigHandler struct {
	logger *logrus.Logger
}

// Updated constructor to accept a Logrus logger
func NewConfigHandler(logger *logrus.Logger) *ConfigHandler {
	return &ConfigHandler{
		logger: logger,
	}
}

// @Summary Update config
// @Description Updates the server's config yaml file
// @ID update-config
// @Accept  json
// @Produce  json
// @Success 200  {string}  string "Config updated successfully"
// @Failure 400  {string}  string "Bad Request"
// @Failure 500  {string}  string "Internal Server Error"
// @Router /api/updateConfig [post]
func (ch *ConfigHandler) UpdateConfig(c *gin.Context) {
	var newConfig config.Config

	if err := c.BindJSON(&newConfig); err != nil {
		ch.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to bind JSON for new configuration")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := config.UpdateConfigFile(newConfig); err != nil {
		ch.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to update configuration file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update configuration"})
		return
	}

	ch.logger.Info("Configuration updated successfully")
	c.Status(http.StatusOK)
}
