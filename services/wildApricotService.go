package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
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
	apiKey := os.Getenv("WILD_APRICOT_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("API key for Wild Apricot is not set in environment variables")
	}

	url := s.TokenEndpoint
	data := "grant_type=client_credentials&scope=auto"
	encodedApiKey := base64.StdEncoding.EncodeToString([]byte("APIKEY:" + apiKey))
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
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
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

// GetContacts retrieves resultId for Contacts request to Wild Apricot for the specified account ID.
func (s *WildApricotService) GetContacts() ([]models.Contact, error) {
	if err := s.refreshTokenIfNeeded(); err != nil {
		log.Printf("Error refreshing token: %v", err)
		return nil, err
	}

	contactURL := fmt.Sprintf("%s/%d/Contacts?$async=false&$filter=%s",
		s.WildApricotApiBase,
		s.cfg.WildApricotAccountId,
		url.QueryEscape(s.cfg.ContactFilterQuery))

	req, err := http.NewRequest("GET", contactURL, nil)
	if err != nil {
		log.Printf("Error creating request for contacts: %v", err)
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+s.ApiToken)
	req.Header.Add("Accept", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		log.Printf("Error during WA contacts fetch: %v", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Unexpected status code %d received", resp.StatusCode)
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	log.Println("WA contacts fetch successful, parsing response")
	return s.parseContactsResponse(resp)
}

// parseContactsResponse parses the HTTP response to extract contact information.
func (s *WildApricotService) parseContactsResponse(resp *http.Response) ([]models.Contact, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading contacts response body: %v", err)
		return nil, err
	}

	var contactsResponse struct {
		Contacts []models.Contact `json:"Contacts"`
	}

	if err = json.Unmarshal(body, &contactsResponse); err != nil {
		log.Printf("Error unmarshalling contacts response: %v", err)
		return nil, err
	}

	log.Printf("Parsed %d contacts from response", len(contactsResponse.Contacts))
	return contactsResponse.Contacts, nil
}
