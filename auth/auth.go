package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"rfid-backend/config"
	"rfid-backend/models"
	"rfid-backend/services"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

type Auth struct {
	cfg       *config.Config
	OAuthConf *oauth2.Config
	Logger    *logrus.Logger
	Store     sessions.Store
	waService *services.WildApricotService
}

func NewAuth(cfg *config.Config, logger *logrus.Logger) *Auth {
	return &Auth{
		waService: services.NewWildApricotService(cfg, logger),
	}
}

func (a *Auth) Initialize(oauthConfig *oauth2.Config, cfg *config.Config, logger *logrus.Logger) {
	a.cfg = cfg
	a.Logger = logger
	a.Logger.Info("Initializing authentication module")

	a.OAuthConf = oauthConfig
	a.OAuthConf.ClientSecret = cfg.SSOClientSecret
	a.Store = cookie.NewStore([]byte(cfg.CookieStoreSecret))

	a.Logger.Info("Authentication module initialized successfully")
}

func (a *Auth) StartOAuthFlow(c *gin.Context) {
	a.Logger.Info("Starting OAuth flow")
	state := a.generateStateOauthToken()
	a.Logger.Infof("Generated OAuth state: %s", state)

	session := sessions.Default(c)
	session.Set("state", state)
	session.Save()

	url := a.OAuthConf.AuthCodeURL(state)
	a.Logger.Infof("Redirecting to: %s", url)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (a *Auth) OAuthCallback(c *gin.Context) {
	a.Logger.Info("Received OAuth callback")
	session := sessions.Default(c)

	receivedState := c.Query("state")
	expectedState := session.Get("state")
	if receivedState != expectedState {
		a.Logger.Errorf("Invalid OAuth state: expected %s, got %s", expectedState, receivedState)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	code := c.Query("code")
	a.Logger.Infof("Exchanging code for token: %s", code)
	if code == "" {
		a.Logger.Error("No code in OAuth callback")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	token, err := a.requestTokenFromWA(code)
	if err != nil {
		a.Logger.WithError(err).Error("Failed to exchange token")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	contact, err := a.waService.GetAuthContact(token)
	if err != nil {
		a.Logger.WithError(err).Error("Failed to extract contact")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	a.Logger.Infof("user id obtained successfully: %s", contact.Id)
	a.handleUserSession(c, contact)
}

func (a *Auth) requestTokenFromWA(code string) (string, error) {
	a.Logger.Info("Preparing data for token request")

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", a.OAuthConf.ClientID)
	data.Set("redirect_uri", a.cfg.SSORedirectURI)
	data.Set("scope", "contacts_me")
	a.Logger.WithFields(logrus.Fields{
		"grant_type":   "authorization_code",
		"code":         code,
		"client_id":    a.OAuthConf.ClientID,
		"redirect_uri": a.cfg.SSORedirectURI,
		"scope":        "contacts_me",
	}).Info("Token request data prepared")

	a.Logger.Info("Creating new HTTP request for token")
	req, err := http.NewRequest("POST", "https://oauth.wildapricot.org/auth/token", strings.NewReader(data.Encode()))
	if err != nil {
		a.Logger.WithError(err).Error("Error creating new HTTP request for token")
		return "", fmt.Errorf("error creating request: %w", err)
	}

	clientCredentials := base64.StdEncoding.EncodeToString([]byte(a.OAuthConf.ClientID + ":" + a.OAuthConf.ClientSecret))
	req.Header.Add("Authorization", "Basic "+clientCredentials)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	a.Logger.Info("HTTP request headers set")

	a.Logger.Info("Sending HTTP request to obtain token")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		a.Logger.WithError(err).Error("Error sending HTTP request to obtain token")
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()
	a.Logger.Info("HTTP request sent, reading response")

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		a.Logger.WithError(err).Error("Error unmarshalling response body")
		return "", fmt.Errorf("error unmarshalling response body: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", errors.New("access token not found in response")
	}

	a.Logger.Info("Access token obtained successfully")
	return tokenResp.AccessToken, nil
}

func (a *Auth) handleUserSession(c *gin.Context, contact *models.Contact) {
	a.Logger.Info("Handling user session")

	if !contact.IsAccountAdministrator {
		c.Redirect(http.StatusFound, "/auth/magicWord")
	}

	session := sessions.Default(c)
	session.Set("user_id", contact.Id)
	session.Set("authenticated", true)
	err := session.Save()
	if err != nil {
		a.Logger.WithError(err).Error("Failed to save session")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	a.Logger.Infof("User session saved for user ID: %s", contact.Id)
	c.Redirect(http.StatusFound, "/web-ui/home")
}

func (a *Auth) generateStateOauthToken() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		a.Logger.Errorf("Error generating state token: %v", err)
		return ""
	}
	token := base64.URLEncoding.EncodeToString(b)
	a.Logger.Infof("Generated state token: %s", token)
	return token
}

func (a *Auth) RequireAuth(c *gin.Context) {
	a.Logger.Info("Checking user authentication")

	session := sessions.Default(c)
	user := session.Get("user_id")
	if user == nil {
		a.Logger.Info("User not authenticated, redirecting to login")
		c.Redirect(http.StatusFound, "/auth/login")
		c.Abort()
	} else {
		a.Logger.Info("User is authenticated")
		c.Next()
	}
}
