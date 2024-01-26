package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"rfid-backend/config"
	"rfid-backend/models"
	"rfid-backend/utils"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type WildApricotService struct {
	Client             *http.Client
	cfg                *config.Config
	TokenEndpoint      string
	ApiToken           string
	WildApricotApiBase string
	TokenExpiry        time.Time
	log                *logrus.Logger
}

var wildApricotSvc = utils.NewSingleton(&WildApricotService{})

func NewWildApricotService(cfg *config.Config, logger *logrus.Logger) *WildApricotService {
	return wildApricotSvc.Get(func() interface{} {
		s := &WildApricotService{
			Client: &http.Client{
				Timeout: time.Second * 30,
			},
			cfg:                cfg,
			TokenEndpoint:      "https://oauth.wildapricot.org/auth/token",
			WildApricotApiBase: "https://api.wildapricot.org/v2/accounts",
			log:                logger,
		}
		s.log.Info("WildApricotService initialized")
		return s
	}).(*WildApricotService)
}

func readResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func handleHTTPError(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
	return nil
}

func (s *WildApricotService) logError(context string, err error) {
	if err != nil {
		s.log.WithFields(logrus.Fields{"context": context, "error": err}).Error("Error occurred")
	}
}

func (s *WildApricotService) buildURL(pathFormat string, args ...interface{}) string {
	return fmt.Sprintf(s.WildApricotApiBase+pathFormat, args...)
}

func unmarshalJSON(body []byte, target interface{}) error {
	return json.Unmarshal(body, target)
}

func (s *WildApricotService) refreshTokenIfNeeded() error {
	if time.Now().After(s.TokenExpiry) || s.ApiToken == "" {
		s.log.Info("Refreshing API token")
		return s.refreshApiToken()
	}
	return nil
}

func (s *WildApricotService) refreshApiToken() error {
	url := s.TokenEndpoint
	data := "grant_type=client_credentials&scope=auto"
	encodedApiKey := base64.StdEncoding.EncodeToString([]byte("APIKEY:" + s.cfg.WildApricotApiKey))
	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		s.logError("Error creating token refresh request: %v", err)
		return err
	}
	req.Header.Add("Authorization", "Basic "+encodedApiKey)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.Client.Do(req)
	if err != nil {
		s.logError("Error during token refresh: %v", err)
		return err
	}

	body, err := readResponseBody(resp)
	if err != nil {
		s.logError("Error reading token response body: %v", err)
		return err
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err = json.Unmarshal(body, &tokenResponse); err != nil {
		s.logError("Error unmarshalling token response: %v", err)
		return err
	}

	expiryDuration := time.Duration(tokenResponse.ExpiresIn) * time.Second
	s.ApiToken = tokenResponse.AccessToken
	s.TokenExpiry = time.Now().Add(expiryDuration)

	s.log.Infof("API token refreshed, expires in: %v", expiryDuration)
	return nil
}

func (s *WildApricotService) makeHTTPRequest(method, url string, body io.Reader, ssoToken string) (*http.Response, error) {
	if len(ssoToken) <= 0 {
		if err := s.refreshTokenIfNeeded(); err != nil {
			s.logError("Error refreshing token: %v", err)
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		s.logError("Error creating HTTP request: %v", err)
		return nil, err
	}
	if len(ssoToken) <= 0 {
		req.Header.Add("Authorization", "Bearer "+s.ApiToken)
	} else {
		req.Header.Add("Authorization", "Bearer "+ssoToken)
	}
	req.Header.Add("Accept", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		s.logError("Error during HTTP request: %v", err)
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

	resp, err := s.makeHTTPRequest("GET", contactURL, nil, "")
	if err != nil {
		s.logError("creating request for contacts", err)
		return nil, err
	}

	if err := handleHTTPError(resp); err != nil {
		s.logError("handling HTTP error for contacts", err)
		return nil, err
	}

	contacts, err := s.parseHTTPResponse(resp)
	if err != nil {
		s.logError("parsing HTTP response", err)
	}

	s.log.Infof("Parsed %d contacts from response", len(contacts))
	return contacts, nil
}

func (s *WildApricotService) GetAuthContact(token string) (*models.Contact, error) {
	contactURL := s.buildURL("/%d/Contacts/me",
		s.cfg.WildApricotAccountId)

	resp, err := s.makeHTTPRequest("GET", contactURL, nil, token)
	if err != nil {
		s.logError("creating request for contact", err)
		return nil, err
	}

	if err := handleHTTPError(resp); err != nil {
		s.logError("handling HTTP error for contact", err)
		return nil, err
	}

	contact, err := s.parseHTTPResponse(resp)
	if err != nil {
		s.logError("parsing HTTP response", err)
		return nil, err
	}

	s.log.Info("Parsed contact from response")
	if len(contact) > 0 {
		return &contact[0], nil
	}

	return nil, fmt.Errorf("no contact found")
}

func (s *WildApricotService) GetContact(contactId int) (*models.Contact, error) {
	contactURL := s.buildURL("/%d/Contacts/%d",
		s.cfg.WildApricotAccountId,
		contactId)

	resp, err := s.makeHTTPRequest("GET", contactURL, nil, "")
	if err != nil {
		s.logError("creating request for contact", err)
		return nil, err
	}

	if err := handleHTTPError(resp); err != nil {
		s.logError("handling HTTP error for contact", err)
		return nil, err
	}

	contact, err := s.parseHTTPResponse(resp)
	if err != nil {
		s.logError("parsing HTTP response", err)
		return nil, err
	}

	s.log.Info("Parsed contact from response")
	if len(contact) > 0 {
		return &contact[0], nil
	}

	return nil, fmt.Errorf("no contact found")
}

func (s *WildApricotService) parseHTTPResponse(resp *http.Response) ([]models.Contact, error) {
	body, err := readResponseBody(resp)
	if err != nil {
		s.logError("Error reading response body: %v", err)
		return nil, err
	}

	// try as multiple contacts
	var contactsResponse struct {
		Contacts []models.Contact `json:"Contacts"`
	}
	if err = json.Unmarshal(body, &contactsResponse); err == nil {
		if len(contactsResponse.Contacts) > 1 {
			s.log.Infof("Parsed %d contacts from response", len(contactsResponse.Contacts))
			return contactsResponse.Contacts, nil
		}
	}
	// try parsing as a single contact
	var contact models.Contact
	if err = json.Unmarshal(body, &contact); err == nil {
		s.log.Info("Parsed single contact from response")
		return []models.Contact{contact}, nil
	}

	// If both attempts fail, return the original unmarshalling error
	s.logError("Error unmarshalling response: %v", err)
	return nil, err
}
