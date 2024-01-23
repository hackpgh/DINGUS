package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"rfid-backend/config"
	"rfid-backend/services"
	"rfid-backend/webhooks"
	"strconv"
)

type WebhooksHandler struct {
	waService *services.WildApricotService
	dbService *services.DBService
	cfg       *config.Config
}

func NewWebhooksHandler(waService *services.WildApricotService, dbService *services.DBService, cfg *config.Config) *WebhooksHandler {
	return &WebhooksHandler{
		waService: waService,
		dbService: dbService,
		cfg:       cfg,
	}
}

func (wh *WebhooksHandler) HandleWebhook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		providedToken := r.URL.Query().Get("token")
		if providedToken != wh.cfg.WildApricotWebhookToken {
			message := fmt.Sprintf("Unauthorized: Invalid token. providedToken=%s, configuration token=%s", providedToken, wh.cfg.WildApricotWebhookToken)
			http.Error(w, message, http.StatusUnauthorized)
			return
		}

		var webhookData webhooks.Webhook
		if err := json.NewDecoder(r.Body).Decode(&webhookData); err != nil {
			http.Error(w, "Failed to decode webhook: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		wh.Process(w, webhookData)
	}
}

// Wild Apricot webhooks will blast multiple webhooks for a single event if a trigger overlap exists.
//
//	Example #1: Admin changes a value in a Contact's Custom Membership Field
//	              -> Contact WA webhook triggers, sending Contact.Id & Action:"Changed", ProfileChanged:"True"
//	                -> Fetch and process new Custom membership Field data
//	                  -> INSERT or DELETE entries in DB members_trainings_link table
//
//	Example #2: Membership Status changes to 'Lapsed'
//	              -> Contact WA webhook triggers, sending Contact.Id, Action:"Changed", ProfileChanged:"False"
//	              -> Membership WA webhook triggers, sending Contact.Id, MembershipStatus, etc.
//	                -> Fetch and process  Custom membership Field for tag data
//	                  -> DELETE entry in DB `members` table
func (wh *WebhooksHandler) Process(w http.ResponseWriter, data webhooks.Webhook) {
	switch data.MessageType {
	case "ContactModified":
		contactParams, ok := data.Parameters.(*webhooks.ContactParameters)
		if !ok {
			http.Error(w, "Invalid contact parameters", http.StatusBadRequest)
			return
		}

		if contactParams.Action == "Changed" && contactParams.ProfileChanged == "True" {
			contact_id, _ := strconv.Atoi(contactParams.ContactId)
			log.Printf("contact_id: %d", contact_id)
			contact, err := wh.waService.GetContact(contact_id)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				log.Printf("Error fetching contact: %v", err)
				return
			}

			if contact == nil {
				log.Println("No contact found for provided ContactID")
			} else {
				wh.dbService.ProcessContactWebhookTrainingData(*contact)
				log.Printf("Webhook notification processed successfully")
			}
		}
	case "Membership":
		membershipParams, ok := data.Parameters.(*webhooks.MembershipParameters)
		if !ok {
			http.Error(w, "Invalid membership parameters", http.StatusBadRequest)
			return
		}

		status := membershipParams.MembershipStatus

		if status != webhooks.StatusNOOP {
			if status == webhooks.StatusLapsed || status == webhooks.StatusActive {
				contact_id, _ := strconv.Atoi(membershipParams.ContactId)
				contact, err := wh.waService.GetContact(contact_id)
				if err != nil {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					log.Printf("Error fetching contact: %v", err)
					return
				}

				wh.dbService.ProcessMembershipWebhook(*membershipParams, *contact)
			}
		}
	default:
		http.Error(w, "Unknown MessageType", http.StatusBadRequest)
		return
	}
}
