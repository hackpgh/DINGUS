package handlers

import (
	"log"
	"net"
	"net/http"
	"rfid-backend/config"
	"rfid-backend/services"

	"github.com/gin-gonic/gin"
)

type RegistrationHandler struct {
	cfg       *config.Config
	dbService *services.DBService
}

func NewRegistrationHandler(dbService *services.DBService, cfg *config.Config) *RegistrationHandler {
	return &RegistrationHandler{
		cfg:       cfg,
		dbService: dbService,
	}
}

func (rh *RegistrationHandler) HandleRegisterDevice(c *gin.Context) {
	deviceName := c.Query("deviceName")
	if deviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Device name is required"})
		return
	}

	ipAddress, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		log.Printf("Failed to get IP address: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get IP address"})
		return
	}

	log.Printf("Registering device %s : %s", deviceName, ipAddress)
	err = rh.dbService.InsertDevice(ipAddress)
	if err != nil {
		log.Printf("Failed to insert device: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert device"})
		return
	}

	// Checking if device requires training
	training, err := rh.dbService.GetTraining(deviceName)
	if err != nil {
		log.Printf("Failed to get training for device: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get training for device"})
		return
	}

	if len(training) > 0 {
		err = rh.dbService.InsertDeviceTrainingLink(ipAddress, training)
		if err != nil {
			log.Printf("Failed to insert device training link: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert device training link"})
			return
		}
	}

	c.Status(http.StatusOK)
}
