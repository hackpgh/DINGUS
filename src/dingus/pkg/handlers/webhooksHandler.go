package handlers

import (
	"dingus/pkg/config"
	"dingus/pkg/services"
	"dingus/pkg/webhooks"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type WebhooksHandler struct {
	waService *services.WildApricotService
	dbService *services.DBService
	cfg       *config.Config
	log       *logrus.Logger
}

func NewWebhooksHandler(waService *services.WildApricotService, dbService *services.DBService, cfg *config.Config, logger *logrus.Logger) *WebhooksHandler {
	return &WebhooksHandler{
		waService: waService,
		dbService: dbService,
		cfg:       cfg,
		log:       logger,
	}
}

// @Summary Handle Wild Apricot webhook requests
// @Description Wild Apricot sends arbitrary JSON per event trigger bsed on their criteria detailed in the official docs
// @ID handle-webhook
// @Accept  json
// @Produce  json
// @Param   token  query    string  true  "Token"
// @Success 200  {string}  string "Webhook processed successfully"
// @Failure 400  {string}  string "Bad Request"
// @Failure 500  {string}  string "Internal Server Error"
// @Router /api/webhooks [post]
func (wh *WebhooksHandler) HandleWebhook(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
		return
	}

	providedToken := c.Query("token")
	if providedToken != wh.cfg.WildApricotWebhookToken {
		message := fmt.Sprintf("Unauthorized: Invalid token. providedToken=%s, configuration token=%s", providedToken, wh.cfg.WildApricotWebhookToken)
		c.JSON(http.StatusUnauthorized, gin.H{"error": message})
		return
	}

	var webhookData webhooks.Webhook
	if err := c.ShouldBindJSON(&webhookData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode webhook: " + err.Error()})
		return
	}

	wh.Process(c, webhookData)
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
func (wh *WebhooksHandler) Process(c *gin.Context, data webhooks.Webhook) {
	switch data.MessageType {
	case "ContactModified":
		wh.handleContactModified(c, data)
	case "Membership":
		wh.handleMembership(c, data)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown MessageType"})
		return
	}

	c.Status(http.StatusOK)
}

func (wh *WebhooksHandler) handleContactModified(c *gin.Context, data webhooks.Webhook) {
	contactParams, ok := data.Parameters.(*webhooks.ContactParameters)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contact parameters"})
		return
	}

	if contactParams.Action == "Changed" && contactParams.ProfileChanged == "True" {
		contactId, _ := strconv.Atoi(contactParams.ContactId)
		wh.log.Infof("contactId: %d", contactId)
		contact, err := wh.waService.GetContact(contactId)
		if err != nil {
			wh.log.Errorf("Error fetching contact: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		if contact == nil {
			wh.log.Error("No contact found for provided ContactID")
		} else {
			wh.dbService.ProcessContactWebhookTrainingData(*contactParams, *contact)
			wh.log.Infof("Webhook notification processed successfully")
		}
	}
}

func (wh *WebhooksHandler) handleMembership(c *gin.Context, data webhooks.Webhook) {
	membershipParams, ok := data.Parameters.(*webhooks.MembershipParameters)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid membership parameters"})
		return
	}

	status := membershipParams.MembershipStatus

	if status != webhooks.StatusNOOP {
		if status == webhooks.StatusLapsed || status == webhooks.StatusActive {
			contactId, _ := strconv.Atoi(membershipParams.ContactId)
			contact, err := wh.waService.GetContact(contactId)
			if err != nil {
				wh.log.Errorf("Error fetching contact: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}

			wh.dbService.ProcessMembershipWebhook(*membershipParams, *contact)
		}
	}
}
