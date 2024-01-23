package handlers

import (
	"log"
	"net"
	"net/http"
	"rfid-backend/config"
	"rfid-backend/services"
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

func (rh *RegistrationHandler) HandleRegisterDevice() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deviceName := r.URL.Query().Get("deviceName")
		if deviceName == "" {
			http.Error(w, "Device name is required", http.StatusBadRequest)
			return
		}

		ipAddress, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Printf("Failed to get IP address: %v", err)
			http.Error(w, "Failed to get IP address", http.StatusInternalServerError)
		}

		log.Printf("Registering device %s : %s", deviceName, ipAddress)
		rh.dbService.InsertDevice(ipAddress)

		// Checking if device requires training
		training, err := rh.dbService.GetTraining(deviceName)
		if err == nil {
			log.Printf("Training is required for the device: %v", err)
			rh.dbService.InsertDeviceTrainingLink(ipAddress, training)
		}

		w.WriteHeader(http.StatusOK)
	}
}
