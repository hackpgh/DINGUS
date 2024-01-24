package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

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
)

// Initialize sets up the necessary configurations for the auth package.
func Initialize(oauthConfig *oauth2.Config, sessionKey []byte, logger *logrus.Logger) {
	OAuthConf = oauthConfig
	Logger = logger
	store := cookie.NewStore(sessionKey)
	gin.Use(sessions.Sessions("mysession", store))
}

// StartOAuthFlow initiates the OAuth2 authentication process.
func StartOAuthFlow(c *gin.Context) {
	state := generateStateOauthToken()
	session := sessions.Default(c)
	session.Set("state", state)
	session.Save()

	url := OAuthConf.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// OAuthCallback handles the OAuth2 callback from the OAuth provider.
func OAuthCallback(c *gin.Context) {
	session := sessions.Default(c)

	// Validate state parameter
	receivedState := c.Query("state")
	expectedState := session.Get("state")
	if receivedState != expectedState {
		Logger.Error("Invalid OAuth state")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Exchange code for access token
	code := c.Query("code")
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

	// Handle user authentication and session management
	handleUserSession(c, token)

	c.Redirect(http.StatusFound, "/web-ui/home")
}

// handleUserSession handles the creation or update of the user session after successful OAuth authentication.
func handleUserSession(c *gin.Context, token *oauth2.Token) {
	// Extract user information from token and create/update session
	// This is an example. Replace with actual logic to handle user session.
	userID := "extracted-user-id" // Replace with actual user ID extraction logic
	session := sessions.Default(c)
	session.Set("user_id", userID)
	session.Set("authenticated", true)
	err := session.Save()
	if err != nil {
		Logger.WithError(err).Error("Failed to save session")
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}

// generateStateOauthToken generates a random state token for OAuth2 flow.
func generateStateOauthToken() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		Logger.Errorf("Error generating state token: %v", err)
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func RequireAuth(c *gin.Context) {
	session := sessions.Default(c)
	if user := session.Get("user_id"); user == nil {
		// Redirect to login if not authenticated
		c.Redirect(http.StatusFound, "/auth/login")
		c.Abort()
	} else {
		c.Next()
	}
}
