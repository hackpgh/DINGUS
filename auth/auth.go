package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"rfid-backend/config"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

var (
	// OAuthConf should be initialized in your main package and passed to auth package.
	OAuthConf *oauth2.Config
	Logger    *logrus.Logger
	store     sessions.Store
)

func Initialize(oauthConfig *oauth2.Config, cfg *config.Config, logger *logrus.Logger) {
	Logger = logger
	Logger.Info("Initializing authentication module")

	OAuthConf = oauthConfig
	OAuthConf.ClientSecret = cfg.SSOClientSecret
	store = cookie.NewStore([]byte(cfg.CookieStoreSecret))

	Logger.Info("Authentication module initialized successfully")
}

func StartOAuthFlow(c *gin.Context) {
	Logger.Info("Starting OAuth flow")
	state := generateStateOauthToken()
	Logger.Infof("Generated OAuth state: %s", state)

	session := sessions.Default(c)
	session.Set("state", state)
	session.Save()

	url := OAuthConf.AuthCodeURL(state)
	Logger.Infof("Redirecting to: %s", url)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func OAuthCallback(c *gin.Context) {
	Logger.Info("Received OAuth callback")
	session := sessions.Default(c)

	receivedState := c.Query("state")
	expectedState := session.Get("state")
	if receivedState != expectedState {
		Logger.Errorf("Invalid OAuth state: expected %s, got %s", expectedState, receivedState)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	code := c.Query("code")
	Logger.Infof("Exchanging code for token: %s", code)
	if code == "" {
		Logger.Error("No code in OAuth callback")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	token, err := OAuthConf.Exchange(c, code)
	if err != nil {
		Logger.WithError(err).Error("Failed to exchange token")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	Logger.Infof("Token exchange successful: %v", token)
	handleUserSession(c, token)
	c.Redirect(http.StatusFound, "/web-ui/home")
}

func handleUserSession(c *gin.Context, token *oauth2.Token) {
	Logger.Info("Handling user session")

	userID := "extracted-user-id" // Replace with actual user ID extraction logic
	session := sessions.Default(c)
	session.Set("user_id", userID)
	session.Set("authenticated", true)
	err := session.Save()
	if err != nil {
		Logger.WithError(err).Error("Failed to save session")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	Logger.Infof("User session saved for user ID: %s", userID)
}

// generateStateOauthToken generates a random state token for OAuth2 flow.
func generateStateOauthToken() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		Logger.Errorf("Error generating state token: %v", err)
		return ""
	}
	token := base64.URLEncoding.EncodeToString(b)
	Logger.Infof("Generated state token: %s", token)
	return token
}

func RequireAuth(c *gin.Context) {
	Logger.Info("Checking user authentication")

	session := sessions.Default(c)
	user := session.Get("user_id")
	if user == nil {
		Logger.Info("User not authenticated, redirecting to login")
		c.Redirect(http.StatusFound, "/auth/login")
		c.Abort()
	} else {
		Logger.Info("User is authenticated")
		c.Next()
	}
}
