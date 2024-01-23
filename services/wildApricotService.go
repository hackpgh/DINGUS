package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"rfid-backend/config"
	"rfid-backend/models"
	"rfid-backend/utils"
	"strings"
	"time"
)

// WildApricotService provides functionalities to interact with the Wild Apricot API.
// It handles the retrieval of contact data and manages API token refresh.
type WildApricotService struct {
	Client             *http.Client
	cfg                *config.Config
	TokenEndpoint      string
	ApiToken           string
	WildApricotApiBase string
	TokenExpiry        time.Time
}

// wildApricotSvc is a singleton instance of WildApricotService.
var wildApricotSvc = utils.NewSingleton(&WildApricotService{})

// NewWildApricotService initializes and retrieves a singleton instance of WildApricotService.
func NewWildApricotService(cfg *config.Config) *WildApricotService {
	return wildApricotSvc.Get(func() interface{} {
		service := &WildApricotService{
			Client: &http.Client{
				Timeout: time.Second * 30,
			},
			cfg:                cfg,
			TokenEndpoint:      "https://oauth.wildapricot.org/auth/token",
			WildApricotApiBase: "https://api.wildapricot.org/v2/accounts",
		}
		log.Println("WildApricotService initialized")
		return service
	}).(*WildApricotService)
}

// readResponseBody reads and returns the body of an HTTP response.
func readResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// handleHTTPError checks for HTTP errors and formats a standard error message.
func handleHTTPError(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
	return nil
}

// logError formats and logs the error messages.
func logError(context string, err error) {
	if err != nil {
		log.Printf("Error %s: %v", context, err)
	}
}

// buildURL constructs and returns a formatted URL string.
func (s *WildApricotService) buildURL(pathFormat string, args ...interface{}) string {
	return fmt.Sprintf(s.WildApricotApiBase+pathFormat, args...)
}

// unmarshalJSON is a utility function to unmarshal JSON into a provided struct.
func unmarshalJSON(body []byte, target interface{}) error {
	return json.Unmarshal(body, target)
}

// refreshTokenIfNeeded checks and refreshes the API token if needed.
func (s *WildApricotService) refreshTokenIfNeeded() error {
	if time.Now().After(s.TokenExpiry) || s.ApiToken == "" {
		log.Println("Refreshing API token")
		return s.refreshApiToken()
	}
	return nil
}

// refreshApiToken handles the token refresh process.
func (s *WildApricotService) refreshApiToken() error {
	url := s.TokenEndpoint
	data := "grant_type=client_credentials&scope=auto"
	encodedApiKey := base64.StdEncoding.EncodeToString([]byte("APIKEY:" + s.cfg.WildApricotApiKey))
	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		log.Printf("Error creating token refresh request: %v", err)
		return err
	}
	req.Header.Add("Authorization", "Basic "+encodedApiKey)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.Client.Do(req)
	if err != nil {
		log.Printf("Error during token refresh: %v", err)
		return err
	}

	body, err := readResponseBody(resp)
	if err != nil {
		log.Printf("Error reading token response body: %v", err)
		return err
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err = json.Unmarshal(body, &tokenResponse); err != nil {
		log.Printf("Error unmarshalling token response: %v", err)
		return err
	}

	expiryDuration := time.Duration(tokenResponse.ExpiresIn) * time.Second
	s.ApiToken = tokenResponse.AccessToken
	s.TokenExpiry = time.Now().Add(expiryDuration)

	log.Printf("API token refreshed, expires in: %v", expiryDuration)
	return nil
}

// makeHTTPRequest handles creating and sending HTTP requests, including token refresh.
func (s *WildApricotService) makeHTTPRequest(method, url string, body io.Reader) (*http.Response, error) {
	if err := s.refreshTokenIfNeeded(); err != nil {
		log.Printf("Error refreshing token: %v", err)
		return nil, err
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Printf("Error creating HTTP request: %v", err)
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+s.ApiToken)
	req.Header.Add("Accept", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		log.Printf("Error during HTTP request: %v", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	return resp, nil
}

func (s *WildApricotService) GetContacts() ([]models.Contact, error) {
	contactURL := s.buildURL("/%d/Contacts?$async=false&$filter=%s",
		s.cfg.WildApricotAccountId,
		url.QueryEscape(s.cfg.ContactFilterQuery))

	resp, err := s.makeHTTPRequest("GET", contactURL, nil)
	if err != nil {
		logError("creating request for contacts", err)
		return nil, err
	}

	if err := handleHTTPError(resp); err != nil {
		logError("handling HTTP error for contacts", err)
		return nil, err
	}

	contacts, err := s.parseHTTPResponse(resp)
	if err != nil {
		logError("parsing HTTP response", err)
	}

	log.Printf("Parsed %d contacts from response", len(contacts))
	return contacts, nil
}

func (s *WildApricotService) GetContact(contactId int) (*models.Contact, error) {
	contactURL := s.buildURL("/%d/Contacts/%d",
		s.cfg.WildApricotAccountId,
		contactId)

	resp, err := s.makeHTTPRequest("GET", contactURL, nil)
	if err != nil {
		logError("creating request for contact", err)
		return nil, err
	}

	if err := handleHTTPError(resp); err != nil {
		logError("handling HTTP error for contact", err)
		return nil, err
	}

	contact, err := s.parseHTTPResponse(resp)
	if err != nil {
		logError("parsing HTTP response", err)
		return nil, err
	}

	log.Printf("Parsed contact from response")
	if len(contact) > 0 {
		return &contact[0], nil
	}

	return nil, fmt.Errorf("no contact found")
}

// parseHTTPResponse parses the HTTP response to extract either a single contact or multiple contacts.
func (s *WildApricotService) parseHTTPResponse(resp *http.Response) ([]models.Contact, error) {
	body, err := readResponseBody(resp)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return nil, err
	}

	// try as multiple contacts
	var contactsResponse struct {
		Contacts []models.Contact `json:"Contacts"`
	}
	if err = json.Unmarshal(body, &contactsResponse); err == nil {
		if len(contactsResponse.Contacts) > 1 {
			log.Printf("Parsed %d contacts from response", len(contactsResponse.Contacts))
			return contactsResponse.Contacts, nil
		}
	}
	// First failure, try parsing as a single contact
	var contact models.Contact
	if err = json.Unmarshal(body, &contact); err == nil {
		log.Printf("Parsed single contact from response")
		return []models.Contact{contact}, nil
	}

	// If both attempts fail, return the original unmarshalling error
	log.Printf("Error unmarshalling response: %v", err)
	return nil, err
}
