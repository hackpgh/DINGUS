package services

import (
	"net/http"
	"net/http/httptest"
	"os"

	"testing"
	"time"

	"rfid-backend/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	mockTokenResponse    = `{"access_token":"mocked_access_token","expires_in":3600}`
	mockContactsResponse = `{
        "Contacts": [
            {
                "FirstName": "John",
                "LastName": "Doe",
                "Email": "test@test.com",
                "DisplayName": "John Doe",
                "Organization": "testorg",
                "ProfileLastUpdated": "2023-01-01T00:00:00",
                "FieldValues": [],
                "Id": 1,
                "Url": "http://test.com/test/1",
                "IsAccountAdministrator": false,
                "TermsOfUseAccepted": true
            }
        ]
    }`
)

func setupMockServer(response string, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte(response))
	}))
}

func TestRefreshApiToken(t *testing.T) {
	mockServer := setupMockServer(mockTokenResponse, http.StatusOK)
	defer mockServer.Close()

	os.Setenv("WILD_APRICOT_API_KEY", "test_api_key")
	defer os.Unsetenv("WILD_APRICOT_API_KEY")

	cfg := &config.Config{}
	service := NewWildApricotService(cfg)
	service.Client = &http.Client{Timeout: time.Second * 30}
	service.TokenEndpoint = mockServer.URL

	err := service.refreshApiToken()
	require.NoError(t, err)
	assert.NotEmpty(t, service.ApiToken)
	assert.True(t, time.Now().Before(service.TokenExpiry))
}

// func TestGetContacts(t *testing.T) {
// 	tokenServer := setupMockServer(mockTokenResponse, http.StatusOK)
// 	contactsServer := setupMockServer(mockContactsResponse, http.StatusOK)
// 	defer tokenServer.Close()
// 	defer contactsServer.Close()

// 	os.Setenv("WILD_APRICOT_API_KEY", "test_api_key")
// 	defer os.Unsetenv("WILD_APRICOT_API_KEY")

// 	cfg := &config.Config{}
// 	service := NewWildApricotService(cfg)
// 	service.Client = &http.Client{Timeout: time.Second * 30}

// 	service.TokenEndpoint = tokenServer.URL
// 	service.WildApricotApiBase = contactsServer.URL
// 	service.cfg.ContactFilterQuery = "test_query"
// 	service.cfg.WildApricotAccountId = 12345
// 	service.TokenExpiry = time.Now().Add(time.Hour)

// 	// Test contact retrieval
// 	contacts, err := service.GetContacts()
// 	require.NoError(t, err)
// 	require.Len(t, contacts, 1)
// 	assert.Equal(t, "John", contacts[0].FirstName)
// 	assert.Equal(t, "Doe", contacts[0].LastName)
// }

// func TestParseContactsResponse(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		responseBody   string
// 		expectedError  bool
// 		expectedLength int
// 	}{
// 		{
// 			name:           "Valid response",
// 			responseBody:   mockContactsResponse,
// 			expectedError:  false,
// 			expectedLength: 1,
// 		},
// 		{
// 			name:           "Empty response",
// 			responseBody:   `{"Contacts":[]}`,
// 			expectedError:  false,
// 			expectedLength: 0,
// 		},
// 		{
// 			name:          "Malformed JSON",
// 			responseBody:  `{"Contacts":[`,
// 			expectedError: true,
// 		},
// 		{
// 			name:          "Incorrect format",
// 			responseBody:  `[]`,
// 			expectedError: true,
// 		},
// 	}
// 	cfg := &config.Config{}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			mockResp := ioutil.NopCloser(bytes.NewReader([]byte(tc.responseBody)))
// 			resp := &http.Response{Body: mockResp}

// 			service := NewWildApricotService(cfg)
// 			contacts, err := service.parseContactsResponse(resp)

// 			if tc.expectedError {
// 				require.Error(t, err)
// 			} else {
// 				require.NoError(t, err)
// 				assert.Equal(t, tc.expectedLength, len(contacts))
// 			}
// 		})
// 	}
// }

// func TestFetchAsyncContacts(t *testing.T) {
// 	// Mock a server to simulate different responses based on the 'resultId' parameter
// 	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		resultId := r.URL.Query().Get("resultId")
// 		switch resultId {
// 		case "success":
// 			w.WriteHeader(http.StatusOK)
// 			w.Write([]byte(`{"Contacts":[{"Id":1,"FirstName":"John","LastName":"Doe"}]}`))
// 		case "accepted":
// 			w.WriteHeader(http.StatusAccepted)
// 		case "error":
// 			w.WriteHeader(http.StatusInternalServerError)
// 		}
// 	}))
// 	defer mockServer.Close()

// 	tests := []struct {
// 		name          string
// 		resultId      string
// 		expectedError bool
// 	}{
// 		{
// 			name:          "Successful async fetch",
// 			resultId:      "success",
// 			expectedError: false,
// 		},
// 		{
// 			name:          "Async fetch with retries",
// 			resultId:      "accepted",
// 			expectedError: false,
// 		},
// 		{
// 			name:          "Error during fetch",
// 			resultId:      "error",
// 			expectedError: true,
// 		},
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			service := NewWildApricotService(nil)
// 			service.Client = &http.Client{Timeout: time.Second * 30}
// 			service.WildApricotApiBase = mockServer.URL

// 			_, err := service.fetchAsyncContacts(tc.resultId)

// 			if tc.expectedError {
// 				require.Error(t, err)
// 			} else {
// 				require.NoError(t, err)
// 			}
// 		})
// 	}
// }
