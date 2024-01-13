package services

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"rfid-backend/models"
	"rfid-backend/utils"
	"strings"
	"time"
)

// WildApricotService provides functionalities to interact with the Wild Apricot API.
// It handles the retrieval of contact data and manages API token refresh.
type WildApricotService struct {
	Client      *http.Client
	ApiToken    string
	TokenExpiry time.Time
}

// wildApricotSvc is a singleton instance of WildApricotService.
var wildApricotSvc = utils.NewSingleton(&WildApricotService{})

// NewWildApricotService initializes and retrieves a singleton instance of WildApricotService.
// It requires a database connection as a dependency.
func NewWildApricotService(database *sql.DB) *WildApricotService {
	return wildApricotSvc.Get(func() interface{} {
		service := &WildApricotService{
			Client: &http.Client{
				Timeout: time.Second * 30,
			},
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

	url := "https://oauth.wildapricot.org/auth/token"
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

// GetContacts retrieves contact data from Wild Apricot for the specified account ID.
func (s *WildApricotService) GetContacts(accountID int) ([]models.Contact, error) {
	if err := s.refreshTokenIfNeeded(); err != nil {
		log.Printf("Error refreshing token: %v", err)
		return nil, err
	}

	url := fmt.Sprintf("https://api.wildapricot.org/v2/accounts/%d/Contacts", accountID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error creating request for contacts: %v", err)
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+s.ApiToken)
	req.Header.Add("Accept", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		log.Printf("Error making request to WildApricot API: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		var asyncResponse struct {
			ResultUrl string `json:"ResultUrl"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&asyncResponse); err != nil {
			log.Printf("Error decoding async response: %v", err)
			return nil, err
		}
		return s.fetchAsyncContacts(asyncResponse.ResultUrl)
	}

	return s.parseContactsResponse(resp)
}

// fetchAsyncContacts handles the retrieval of contacts from an async response (WA supports optional async=false request param).
func (s *WildApricotService) fetchAsyncContacts(resultUrl string) ([]models.Contact, error) {
	if err := s.refreshTokenIfNeeded(); err != nil {
		log.Printf("Error refreshing token for async contacts fetch: %v", err)
		return nil, err
	}

	maxRetries := 10 // Maximum number of polling attempts
	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequest("GET", resultUrl, nil)
		if err != nil {
			log.Printf("Error creating request for async contacts: %v", err)
			return nil, err
		}
		req.Header.Add("Authorization", "Bearer "+s.ApiToken)
		req.Header.Add("Accept", "application/json")

		resp, err := s.Client.Do(req)
		if err != nil {
			log.Printf("Error during async contacts fetch: %v", err)
			return nil, err
		}

		// Handling different status codes
		switch resp.StatusCode {
		case http.StatusOK:
			log.Println("Async contacts fetch successful, parsing response")
			return s.parseContactsResponse(resp)
		case http.StatusAccepted:
			// Continue polling
			log.Println("Waiting for async contacts response, polling...")
			time.Sleep(5 * time.Second)
		default:
			// Handle other unexpected status codes
			log.Printf("Unexpected status code %d received", resp.StatusCode)
			return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
		}
		resp.Body.Close() // Ensure the response body is closed after each request
	}

	return nil, fmt.Errorf("max retries reached for async contacts fetch")
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
