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

	"rfid-backend/pkg/config"
	"rfid-backend/pkg/services"
	"rfid-backend/pkg/setup"

	"github.com/gin-gonic/gin"
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

func main() {
	log.Print(hackPghBanner)
	cfg := config.LoadConfig()
	logger := setup.SetupLogger(cfg)

	db, err := setup.SetupDatabase(cfg, logger)
	if err != nil {
		logger.Fatalf("Failed to setup database: %v", err)
	}
	defer db.Close()

	router := gin.Default()

	setup.SetupRoutes(router, cfg, db, logger)

	waService := services.NewWildApricotService(cfg, logger)
	dbService := services.NewDBService(db, cfg, logger)

	setup.StartBackgroundDatabaseUpdate(waService, dbService, logger)

	err = router.RunTLS(":443", cfg.CertFile, cfg.KeyFile)
	if err != nil {
		logger.Fatalf("Failed to start HTTPS server: %v", err)
	}
}
