package sso

import (
	// ... other imports ...
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// Assuming you have a store for Gorilla sessions
var store = sessions.NewCookieStore([]byte("your-secret-key"))

type OAuthHandler struct {
	config       *oauth2.Config
	logger       *logrus.Logger
	sessionStore *sessions.CookieStore
}

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	AuthURL      string
	TokenURL     string
}

func NewSSOHandler(oauthConfig *Config, logger *logrus.Logger, store *sessions.CookieStore) *OAuthHandler {
	// Configure OAuth2 with the provided settings
	return &OAuthHandler{
		config: &oauth2.Config{
			ClientID:     oauthConfig.ClientID,
			ClientSecret: oauthConfig.ClientSecret,
			RedirectURL:  oauthConfig.RedirectURL,
			Scopes:       oauthConfig.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  oauthConfig.AuthURL,
				TokenURL: oauthConfig.TokenURL,
			},
		},
		logger:       logger,
		sessionStore: store,
	}
}

func (h *OAuthHandler) StartOAuthFlow(c *gin.Context) {
	state := generateStateOauthToken()
	session, err := h.sessionStore.Get(c.Request, "oauth-state")
	if err != nil {
		h.logger.WithError(err).Error("Error getting OAuth state session")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	session.Values["state"] = state
	err = session.Save(c.Request, c.Writer)
	if err != nil {
		h.logger.WithError(err).Error("Error saving OAuth state session")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	url := h.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	h.logger.Infof("Redirecting user to OAuth login page: %s", url)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *OAuthHandler) OAuthCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		h.logger.Error("OAuth callback received with no code in request URL")
		c.Redirect(http.StatusFound, "/web-ui/login")
		return
	}

	// Create a session and save the token
	session, err := h.sessionStore.Get(c.Request, "oauth-state")
	if err != nil {
		h.logger.WithError(err).Error("Error getting OAuth state session")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if c.Query("state") != session.Values["state"] {
		h.logger.Error("Invalid OAuth state")
		c.Redirect(http.StatusFound, "/web-ui/login")
		return
	}

	session.Values["authenticated"] = true
	// Save other needed values in session
	// session.Values["user_id"] = "some-user-id"

	err = session.Save(c.Request, c.Writer)
	if err != nil {
		h.logger.WithError(err).Error("Error saving session")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving session"})
		return
	}

	c.Redirect(http.StatusFound, "/web-ui/home")
}

func (h *OAuthHandler) RequireAuth(c *gin.Context) {
	session, err := h.sessionStore.Get(c.Request, "session-name")
	if err != nil || session.Values["authenticated"] != true {
		h.logger.Warn("No valid session found or user not authenticated")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Continue down the chain if authenticated
	c.Next()
}

func generateStateOauthToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
