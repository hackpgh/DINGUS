package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"rfid-backend/config"
	"rfid-backend/services"
)

// CacheHandler handles requests related to RFID and machine cache.
// It interacts with the database service to fetch and provide the requested db data.
type CacheHandler struct {
	cfg       *config.Config      // Configuration settings
	dbService *services.DBService // Database service for data retrieval
}

// NewCacheHandler creates a new instance of CacheHandler.
func NewCacheHandler(dbService *services.DBService, cfg *config.Config) *CacheHandler {
	return &CacheHandler{
		cfg:       cfg,
		dbService: dbService,
	}
}

// HandleMachineCacheRequest creates an HTTP handler function that responds with RFID tags trained for a specific machine.
// It expects a 'machineName' query parameter in the request.
func (ch *CacheHandler) HandleMachineCacheRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		machineName := r.URL.Query().Get("machineName")
		if machineName == "" {
			http.Error(w, "Machine name is required", http.StatusBadRequest)
			return
		}

		log.Printf("Fetching RFID tags for machine: %s", machineName)
		rfids, err := ch.dbService.GetRFIDsForMachine(machineName)
		if err != nil {
			log.Printf("Error fetching RFID tags for machine %s: %v", machineName, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(rfids)
	}
}

// HandleDoorCacheRequest creates an HTTP handler function that responds with all RFID tags.
// It is used to manage access control for doors.
func (ch *CacheHandler) HandleDoorCacheRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Fetching all RFID tags")
		rfids, err := ch.dbService.GetAllRFIDs()
		if err != nil {
			log.Printf("Error fetching all RFID tags: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(rfids)
	}
}
