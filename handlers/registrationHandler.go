package handlers

import (
	"io"
	"net"
	"net/http"
	"rfid-backend/config"
	"rfid-backend/services"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type DeviceWithTraining struct {
	IPAddress        string
	MACAddress       string
	SelectedTraining string
}

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
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		rh.log.Errorf("Failed to get IP address: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get IP address"})
		return
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		rh.log.Errorf("Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read MAC address from request body"})
		return
	}
	macAddress := string(bodyBytes)

	if macAddress == "" {
		rh.log.Errorf("MAC address is empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "MAC address is required"})
		return
	}

	rh.log.Infof("Registering device with IP %s and MAC %s", ip, macAddress)

	err = rh.dbService.InsertDevice(ip, macAddress, 0)
	if err != nil {
		rh.log.Errorf("Failed to insert device: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert device"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device registered successfully"})
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
	trainings = append(trainings, "Door") // Manually append Door since it is not a device that requires training i.e. not on the DB

	dtl, err := rh.dbService.GetDevicesTrainings()
	if err != nil {
		rh.log.Errorf("Failed to get device mappings: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get trainings"})
		return
	}

	macToTraining := make(map[string]string)
	for _, dt := range dtl {
		macToTraining[dt.MACAddress] = dt.Label
	}

	var devicesWithLabels []DeviceWithTraining
	for _, device := range devices {
		selectedTraining := macToTraining[device.MACAddress]
		devicesWithLabels = append(devicesWithLabels, DeviceWithTraining{
			IPAddress:        device.IPAddress,
			MACAddress:       device.MACAddress,
			SelectedTraining: selectedTraining,
		})
	}

	c.HTML(http.StatusOK, "deviceManagement.tmpl", gin.H{
		"DevicesWithLabels": devicesWithLabels,
		"Trainings":         trainings,
	})
}

type DeviceAssignment struct {
	IPAddress     string `json:"ipAddress"`
	MACAddress    string `json:"macAddress"`
	TrainingLabel string `json:"trainingLabel"`
}

func (rh *RegistrationHandler) UpdateDeviceAssignments(c *gin.Context) {
	var assignments []DeviceAssignment

	// Bind the incoming JSON payload to the assignments slice
	if err := c.BindJSON(&assignments); err != nil {
		rh.log.Errorf("Failed to bind JSON data: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON data"})
		return
	}

	// Iterate over each assignment and process it
	for _, assignment := range assignments {
		rh.log.Infof("Processing assignment for device %s", assignment.MACAddress)

		// Determine if a training label indicates a special condition (e.g., "door")
		trainingRequired := strings.Contains(strings.ToLower(assignment.TrainingLabel), "door")

		// Insert or update the device with its training label as needed
		if trainingRequired {
			err := rh.dbService.InsertDevice(assignment.IPAddress, assignment.MACAddress, 1)
			if err != nil {
				rh.log.Errorf("Failed to insert device assignment for %s: %v", assignment.MACAddress, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update device assignments"})
				return
			}
		} else {
			err := rh.dbService.InsertDevice(assignment.IPAddress, assignment.MACAddress, 0)
			if err != nil {
				rh.log.Errorf("Failed to insert device assignment for %s: %v", assignment.MACAddress, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update device assignments"})
				return
			}
		}
		err := rh.dbService.InsertDeviceTrainingLink(assignment.MACAddress, assignment.TrainingLabel)
		if err != nil {
			rh.log.Errorf("Failed to process device assignment for %s: %v", assignment.MACAddress, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update device assignments"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device assignments updated successfully"})
}
