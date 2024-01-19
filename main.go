/*
RFID Backend Server
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

	//_ "net/http/pprof"
	"rfid-backend/config"
	"rfid-backend/db"
	"rfid-backend/handlers"
	"rfid-backend/services"
	"time"

	"github.com/joho/godotenv"
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
| RFID Backend Server for HackPGH                                  |
| - Configure via 'config.yml' or '/' endpoint                     |
| - Serves '/api/doorCache' & '/api/machineCache?machineName= '    |
| - Ensure SSL certificates are in place for HTTPS.                |
| Want to contribute? https://github.com/hackpgh/rfid-backend      |
+------------------------------------------------------------------+
`

func main() {
	// // pprof monitoring, import '_ /net/http/pprof'
	// go func() {
	// 	_ = http.ListenAndServe("0.0.0.0:8081", nil)
	// }()

	log.Print(hackPghBanner)
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg := config.LoadConfig()
	log.Printf("Certificate File: %s", cfg.CertFile)
	log.Printf("Key File: %s", cfg.KeyFile)
	log.Printf("Database Path: %s", cfg.DatabasePath)
	db, err := db.InitDB(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	wildApricotSvc := services.NewWildApricotService(cfg)
	dbService := services.NewDBService(db, cfg)

	configHandler := handlers.NewConfigHandler()

	webhooksHandler := handlers.NewWebhooksHandler(wildApricotSvc, dbService)

	// Configuration web-ui endpoint
	// TODO: Add auth level restriction that allows access to members only if they have sufficient membership level
	http.Handle("/", http.FileServer(http.Dir("web-ui")))

	// Configuration management endpoints
	http.HandleFunc("/api/getConfig", configHandler.GetConfig)
	http.HandleFunc("/api/updateConfig", configHandler.UpdateConfig)

	// Access Control system tags data endpoints for rfid readers' cache
	http.HandleFunc("/api/webhooks", webhooksHandler.HandleWebhook())

	// Start background task to fetch contacts and update the database
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		for range ticker.C {
			updateDatabaseFromWildApricot(wildApricotSvc, dbService)
		}
	}()

	log.Println("Starting HTTPS server on :443...")
	err = http.ListenAndServeTLS(":443", cfg.CertFile, cfg.KeyFile, nil)
	if err != nil {
		log.Fatalf("Failed to start HTTPS server: %v", err)
	}
}

func updateDatabaseFromWildApricot(waService *services.WildApricotService, dbService *services.DBService) {
	log.Println("Fetching contacts from Wild Apricot and updating database...")
	contacts, err := waService.GetContacts()
	if err != nil {
		log.Printf("Failed to fetch contacts: %v", err)
		return
	}

	if len(contacts) <= 0 {
		log.Println("No contacts to process from Wild Apricot. Sleeping...")
		return
	}

	if err = dbService.ProcessContactsData(contacts); err != nil {
		log.Printf("Failed to update database: %v", err)
		return
	} else {
		log.Println("Latest Wild Apricot contacts successfully processed.")
	}

}
