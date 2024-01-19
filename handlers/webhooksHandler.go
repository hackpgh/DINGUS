package handlers

import (
	// ... other imports ...
	"encoding/json"
	"log"
	"net/http"
	"rfid-backend/models"
	"rfid-backend/services"
)

type WebhooksHandler struct {
	waService *services.WildApricotService
	dbService *services.DBService
}

func NewWebhooksHandler(waService *services.WildApricotService, dbService *services.DBService) *WebhooksHandler {
	return &WebhooksHandler{
		waService: waService,
		dbService: dbService,
	}
}

func (s *WebhooksHandler) HandleWebhook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var webhookData models.WildApricotWebhook
		if err := json.NewDecoder(r.Body).Decode(&webhookData); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Process the webhook data
		// For example, update the database based on the received webhook data
		contact, err := s.waService.GetContact(webhookData.Parameters.ContactId)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if len(contact) <= 0 {
			log.Println("No contact found for provided ContactID")
		} else {
			s.dbService.ProcessContactsData(contact)
			log.Printf("Webhook notification processed successfully")
		}

	}
}
