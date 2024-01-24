/*
DINGUS
===================
This Go server provides backend services for an RFID-based access control system
for HackPGH's magnetic locks and machines that require Safety Training sign-offs.
It interacts with the Wild Apricot API to fetch contact data and manages
a local SQLite database to store and process RFID tags and training information.

Project Structure:
- /config: Configuration file loading logic.
- /db: Database initialization and schema management.
- /db/schema: Database schema files.
- /handlers: HTTP handlers for different server endpoints.
- /models: Data structures representing database entities and API responses.
- /services: Business logic, including interaction with external APIs
             and database operations; also contains queries.go

Main Functionality:
- Initializes the SQLite database using the specified database path from `config.yml`.
- Sets up the Wild Apricot service for API interactions, enabling the retrieval of contact data.
- Creates a DBService instance for handling database operations.
- Initializes a CacheHandler with the DBService and configuration settings to handle HTTP requests.
- Registers HTTP endpoints `/api/machineCache` and `/api/doorCache` for fetching RFID data
  related to machines and door access.
- Starts a background routine that periodically fetches contact data from the Wild Apricot
  API and updates the local SQLite database. This ensures the database is regularly
  synchronized with the latest data from Wild Apricot.
- Launches an HTTPS server on port 443 to listen for incoming requests, using the SSL
  certificate and key specified in the `config.yml`.

Usage:
- Before running, ensure that the `config.yml` is properly set up with the necessary configuration, including database path, Wild Apricot account ID, SSL certificate, and key file locations.
- Run the server to start listening for HTTP requests on port 443 and to keep the local database synchronized with the Wild Apricot API data.
*/

package main

import (
	"log"
	"net/http"
	"time"

	"rfid-backend/auth"
	"rfid-backend/config"
	"rfid-backend/db"
	"rfid-backend/handlers"
	"rfid-backend/services"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const hackPghBanner = `
+------------------------------------------------------------------+
|  __    __                   __       _______   ______  __    __  |
| |  \  |  \                 |  \     |       \ /      \|  \  |  \ |
| | ▓▓  | ▓▓ ______   _______| ▓▓   __| ▓▓▓▓▓▓▓\  ▓▓▓▓▓▓\ ▓▓  | ▓▓ |
| | ▓▓__| ▓▓|      \ /       \ ▓▓  /  \ ▓▓__/ ▓▓ ▓▓ __\▓▓ ▓▓__| ▓▓ |
| | ▓▓    ▓▓ \▓▓▓▓▓▓\  ▓▓▓▓▓▓▓ ▓▓_/  ▓▓ ▓▓    ▓▓ ▓▓|    \ ▓▓    ▓▓ |
| | ▓▓▓▓▓▓▓▓/      ▓▓ ▓▓     | ▓▓   ▓▓| ▓▓▓▓▓▓▓| ▓▓ \▓▓▓▓ ▓▓▓▓▓▓▓▓ |
| | ▓▓  | ▓▓  ▓▓▓▓▓▓▓ ▓▓_____| ▓▓▓▓▓▓\| ▓▓     | ▓▓__| ▓▓ ▓▓  | ▓▓ |
| | ▓▓  | ▓▓\▓▓    ▓▓\▓▓     \ ▓▓  \▓▓\ ▓▓      \▓▓    ▓▓ ▓▓  | ▓▓ |
|  \▓▓   \▓▓ \▓▓▓▓▓▓▓ \▓▓▓▓▓▓▓\▓▓   \▓▓\▓▓       \▓▓▓▓▓▓ \▓▓   \▓▓ |
|                                                                  |
|                   Be Excellent to Each Other                     |
+------------------------------------------------------------------+
| DINGUS for HackPGH                                  |
| - Configure via 'config.yml' or '/' endpoint                     |
| - Serves '/api/doorCache' & '/api/machineCache?machineName= '    |
| - Ensure SSL certificates are in place for HTTPS.                |
| Want to contribute? https://github.com/hackpgh/rfid-backend      |
+------------------------------------------------------------------+
`

var oauthConf = &oauth2.Config{
	ClientID:     "your-client-id",
	ClientSecret: "your-client-secret",
	RedirectURL:  "https://yourapp.com/auth/callback",
	Scopes:       []string{"scope1", "scope2"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "provider-auth-url",
		TokenURL: "provider-token-url",
	},
}

func main() {
	// // pprof monitoring, import '_ /net/http/pprof'
	//
	//	go func() {
	//		_ = http.ListenAndServe("0.0.0.0:8081", nil)
	//	}()
	log.Print(hackPghBanner)

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg := config.LoadConfig()

	db, err := db.InitDB(cfg.DatabasePath)
	if err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	router := gin.Default()
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))

	waService := services.NewWildApricotService(cfg, logger)
	dbService := services.NewDBService(db, cfg, logger)

	// Set up OAuth routes
	auth.Initialize(oauthConf, logger)
	authGroup := router.Group("/auth")
	{
		authGroup.GET("/login", auth.StartOAuthFlow)
		authGroup.GET("/callback", auth.OAuthCallback)
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
	logger.Info("Static files are set up")
	router.LoadHTMLGlob("web-ui/templates/*")

	webUI := router.Group("/web-ui")
	{
		webUI.GET("/home", func(c *gin.Context) {
			logger.Info("Serving the home page")
			c.HTML(http.StatusOK, "home.tmpl", nil)
			if c.Writer.Status() == http.StatusOK {
				logger.Info("Home page rendered successfully")
			} else {
				logger.Errorf("Failed to render the home page, status code: %d", c.Writer.Status())
			}
		})
		webUI.GET("/configManagement", func(c *gin.Context) {
			c.HTML(http.StatusOK, "configManagement.tmpl", gin.H{"title": "Configuration Management", "head": `	<link href=\"/css/configManagement.css\" rel=\"stylesheet\">`})
		})
		webUI.GET("/deviceManagement", func(c *gin.Context) {
			c.HTML(http.StatusOK, "deviceManagement.tmpl", gin.H{"title": "Device Management", "head": `	<link href=\"/css/deviceManagement.css\" rel=\"stylesheet\">`})
		})
	}

	webUI.Use(auth.StartOAuthFlow)

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"message": "Not Found"})
	})

	go backgroundDatabaseUpdate(waService, dbService, logger)

	logger.Infof("Gin mode: %s", gin.Mode())
	logger.Info("Starting HTTPS server on :443...")
	err = router.RunTLS(":443", cfg.CertFile, cfg.KeyFile)
	if err != nil {
		logger.Fatalf("Failed to start HTTPS server: %v", err)
	}
}

func GinLogrus(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		logger.WithFields(logrus.Fields{
			"status_code": c.Writer.Status(),
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"ip":          c.ClientIP(),
			"user_agent":  c.Request.UserAgent(),
			"latency":     duration,
		}).Info("handled request")
	}
}

func backgroundDatabaseUpdate(waService *services.WildApricotService, dbService *services.DBService, logger *logrus.Logger) {
	// Run full database sync on startup then repeat on ticker interval
	updateEntireDatabaseFromWildApricot(waService, dbService, logger)

	ticker := time.NewTicker(30 * time.Minute)
	for range ticker.C {
		updateEntireDatabaseFromWildApricot(waService, dbService, logger)
	}
}

func updateEntireDatabaseFromWildApricot(waService *services.WildApricotService, dbService *services.DBService, logger *logrus.Logger) {
	logger.Info("Fetching contacts from Wild Apricot and updating database...")
	contacts, err := waService.GetContacts()
	if err != nil {
		logger.Errorf("Failed to fetch contacts: %v", err)
		return
	}

	if len(contacts) <= 0 {
		logger.Info("No contacts to process from Wild Apricot. Sleeping...")
		return
	}

	if err = dbService.ProcessContactsData(contacts); err != nil {
		logger.Errorf("Failed to update database: %v", err)
		return
	} else {
		logger.Info("Latest Wild Apricot contacts successfully processed.")
	}

}
