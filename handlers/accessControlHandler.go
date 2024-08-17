package handlers

import (
	"net/http"
	"rfid-backend/services"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AccessControlHandler struct {
	dbService *services.DBService
	log       *logrus.Logger
}

func NewAccessControlHandler(dbService *services.DBService, logger *logrus.Logger) *AccessControlHandler {
	return &AccessControlHandler{
		dbService: dbService,
		log:       logger,
	}
}

// @Summary Authenticate a tag swipe
// @Description Authenticates a tag swipe against the db
// @ID authenticate
// @Accept  json
// @Produce  json
// @Success 200  {string}  string "Device registered successfully"
// @Failure 400  {string}  string "Bad Request"
// @Failure 500  {string}  string "Internal Server Error"
// @Router /api/authenticate [post]
func (ach *AccessControlHandler) HandleAuthenticate(c *gin.Context) {
	// Read the raw data from the request body
	data, err := c.GetRawData()
	if err != nil {
		ach.log.Printf("Error reading request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request"})
		return
	}

	// Convert the data to a string and trim any whitespace
	tag := strings.TrimSpace(string(data))

	if tag == "" {
		ach.log.Println("Authentication request missing tag data")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing tag data"})
		return
	}

	// Log the received tag for debugging purposes
	ach.log.Printf("Received tag for verification: %s", tag)

	// Proceed with tag verification...
	if ach.authenticateTag(tag) {
		c.Status(http.StatusOK)
	} else {
		c.Status(http.StatusUnauthorized)
	}
}

func (ach *AccessControlHandler) authenticateTag(raw_tag string) bool {
	exists, err := ach.dbService.TagExists(raw_tag)
	if err != nil {
		ach.log.Printf("Error checking tag existence: %v", err)
		return false
	}
	return exists
}
