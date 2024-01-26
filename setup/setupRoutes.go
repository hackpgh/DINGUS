package setup

import (
	"database/sql"
	"net/http"
	"rfid-backend/auth"
	"rfid-backend/config"
	"rfid-backend/handlers"
	"rfid-backend/services"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/oauth2"
)

func SetupRoutes(router *gin.Engine, cfg *config.Config, db *sql.DB, logger *logrus.Logger) {
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))

	waService := services.NewWildApricotService(cfg, logger)
	dbService := services.NewDBService(db, cfg, logger)

	oauthConf := &oauth2.Config{
		ClientID:     cfg.SSOClientID,
		ClientSecret: cfg.SSOClientSecret,
		RedirectURL:  cfg.SSORedirectURI,
		Scopes:       []string{"contacts_me"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://davidbouw36728.wildapricot.org/sys/login/OAuthLogin",
			TokenURL: "https://oauth.wildapricot.org/auth/token",
		},
	}
	auth := auth.NewAuth(cfg, logger)
	auth.Initialize(oauthConf, cfg, logger)

	authGroup := router.Group("/auth")
	{
		authGroup.GET("/login", auth.StartOAuthFlow)
		authGroup.GET("/callback", auth.OAuthCallback)
		authGroup.GET("/magicWord", func(c *gin.Context) {
			c.HTML(http.StatusOK, "magicWord.tmpl", nil)
		})
	}

	url := ginSwagger.URL("https://localhost:443/swagger/doc.json")
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	registrationHandler := handlers.NewRegistrationHandler(dbService, cfg, logger)
	api := router.Group("/api")
	{
		webhooksHandler := handlers.NewWebhooksHandler(waService, dbService, cfg, logger)
		configHandler := handlers.NewConfigHandler(logger)

		api.POST("/updateConfig", configHandler.UpdateConfig)
		api.POST("/webhooks", webhooksHandler.HandleWebhook)
		api.POST("/registerDevice", registrationHandler.HandleRegisterDevice)
		api.POST("/updateDeviceAssignments", registrationHandler.UpdateDeviceAssignments)
	}

	router.Static("/css", "./web-ui/css")
	router.Static("/js", "./web-ui/js")
	router.Static("/assets", "./web-ui/assets")
	router.LoadHTMLGlob("web-ui/templates/*")

	setupWebUIRoutes(router, logger, auth)
}

func setupWebUIRoutes(router *gin.Engine, logger *logrus.Logger, auth *auth.Auth) {
	webUI := router.Group("/web-ui")
	{
		webUI.Use(auth.RequireAuth)
		webUI.GET("/home", func(c *gin.Context) {
			logger.Info("Serving the home page")
			c.HTML(http.StatusOK, "home.tmpl", nil)
		})
		webUI.GET("/configManagement", func(c *gin.Context) {
			c.HTML(http.StatusOK, "configManagement.tmpl", gin.H{"title": "Configuration Management"})
		})
		webUI.GET("/deviceManagement", func(c *gin.Context) {
			c.HTML(http.StatusOK, "deviceManagement.tmpl", gin.H{"title": "Device Management"})
		})
		router.GET("/web-ui/members", func(c *gin.Context) {
			c.HTML(http.StatusOK, "members.tmpl", gin.H{
				"title": "Wild Apricot Members Portal",
			})
		})
	}
}
