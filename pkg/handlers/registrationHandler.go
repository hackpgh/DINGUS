package handlers

import (
	"net"
	"net/http"
	"rfid-backend/pkg/config"
	"rfid-backend/pkg/services"
	"strings"

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
// @Param   trainingLabel  query    string  true  "Device Name"
// @Success 200  {string}  string "Device registered successfully"
// @Failure 400  {string}  string "Bad Request"
// @Failure 500  {string}  string "Internal Server Error"
// @Router /api/registerDevice [post]
func (rh *RegistrationHandler) HandleRegisterDevice(c *gin.Context) {
	trainingLabel := c.Query("trainingLabel")
	if trainingLabel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Device name is required"})
		return
	}

	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		rh.log.Errorf("Failed to get IP address: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get IP address"})
		return
	}

	rh.log.Infof("Registering device %s : %s", trainingLabel, ip)
	switch {
	case strings.Contains(strings.ToLower(trainingLabel), "door"):
		err := rh.dbService.InsertDevice(ip, 0) // SQLite uses integer 0 & 1 instead of a bool type
		if err != nil {
			rh.log.Errorf("Failed to insert device: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert device"})
			return
		}
	default:
		err := rh.dbService.InsertDevice(ip, 1)
		if err != nil {
			rh.log.Errorf("Failed to insert device: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert device"})
			return
		}
	}

	// Checking if device requires training
	training, err := rh.dbService.GetTraining(trainingLabel)
	if err != nil {
		rh.log.Errorf("Failed to get training for device: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get training for device"})
		return
	}

	if len(training) > 0 {
		err = rh.dbService.InsertDeviceTrainingLink(ip, training)
		if err != nil {
			rh.log.Errorf("Failed to insert device training link: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert device training link"})
			return
		}
	}

	c.Status(http.StatusOK)
}

// @Summary Serve Device Management Page
// @Description Serves the page for managing device and training assignments.
// @ID serve-device-management-page
// @Produce html
// @Success 200 {string} string "Page served successfully"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/deviceManagement [get]
func (rh *RegistrationHandler) ServeDeviceManagementPage(c *gin.Context) {
	devices, err := rh.dbService.GetDevices()
	if err != nil {
		rh.log.Errorf("Failed to get devices: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get devices"})
		return
	}

	trainings, err := rh.dbService.GetAllTrainings()
	if err != nil {
		rh.log.Errorf("Failed to get trainings: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get trainings"})
		return
	}

	c.HTML(http.StatusOK, "deviceManagement.html", gin.H{
		"Devices":   devices,
		"Trainings": trainings,
	})
}

func (rh *RegistrationHandler) UpdateDeviceAssignments(c *gin.Context) {
	formData := make(map[string]string)

	if err := c.Bind(formData); err != nil {
		rh.log.Errorf("Failed to bind form data: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data"})
		return
	}

	for ip, trainingLabel := range formData {

		rh.log.Infof("Inserting Device %s", ip)
		switch {
		case strings.Contains(strings.ToLower(trainingLabel), "door"):
			err := rh.dbService.InsertDevice(ip, 0) // SQLite uses integer 0 & 1 instead of a bool type
			if err != nil {
				rh.log.Errorf("Failed to insert device: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert device"})
				return
			}
		default:
			rh.log.Infof("Training required detected")
			err := rh.dbService.InsertDevice(ip, 1)
			if err != nil {
				rh.log.Errorf("Failed to insert device: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert device"})
				return
			}
		}

		rh.log.Infof("Updating Device Assignment %s:%s", ip, trainingLabel)
		if trainingLabel != "" {
			if err := rh.dbService.InsertDeviceTrainingLink(ip, trainingLabel); err != nil {
				rh.log.Errorf("Failed to update device training link for device %s: %v", ip, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update device assignments"})
				return
			}
		} else {
			if err := rh.dbService.DeleteDeviceTrainingLink(ip); err != nil {
				rh.log.Errorf("Failed to delete device training link for device %s: %v", ip, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update device assignments"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device assignments updated successfully"})
}
