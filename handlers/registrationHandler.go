package handlers

import (
	"net"
	"net/http"
	"rfid-backend/config"
	"rfid-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type RegistrationHandler struct {
	cfg       *config.Config
	dbService *services.DBService
	log       *logrus.Logger
}

func NewRegistrationHandler(dbService *services.DBService, cfg *config.Config, logger *logrus.Logger) *RegistrationHandler {
	return &RegistrationHandler{
		cfg:       cfg,
		dbService: dbService,
		log:       logger,
	}
}

// @Summary Register device
// @Description Register a new device with its IP and name.
// @ID register-device
// @Accept  json
// @Produce  json
// @Param   deviceName  query    string  true  "Device Name"
// @Success 200  {string}  string "Device registered successfully"
// @Failure 400  {string}  string "Bad Request"
// @Failure 500  {string}  string "Internal Server Error"
// @Router /api/registerDevice [post]
func (rh *RegistrationHandler) HandleRegisterDevice(c *gin.Context) {
	deviceName := c.Query("deviceName")
	if deviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Device name is required"})
		return
	}

	ipAddress, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		rh.log.Errorf("Failed to get IP address: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get IP address"})
		return
	}

	rh.log.Infof("Registering device %s : %s", deviceName, ipAddress)
	err = rh.dbService.InsertDevice(ipAddress)
	if err != nil {
		rh.log.Errorf("Failed to insert device: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert device"})
		return
	}

	// Checking if device requires training
	training, err := rh.dbService.GetTraining(deviceName)
	if err != nil {
		rh.log.Errorf("Failed to get training for device: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get training for device"})
		return
	}

	if len(training) > 0 {
		err = rh.dbService.InsertDeviceTrainingLink(ipAddress, training)
		if err != nil {
			rh.log.Errorf("Failed to insert device training link: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert device training link"})
			return
		}
	}

	c.Status(http.StatusOK)
}
