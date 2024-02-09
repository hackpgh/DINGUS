// file: /src/dingus/main.go
package main

import (
	"log"

	"dingus/pkg/config"
	"dingus/pkg/services"
	"dingus/pkg/setup"

	"github.com/gin-gonic/autotls"
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
| DINGUS for HackPGH                                               |
| - Configure via 'config.yaml' or '/web-ui/configManagement'      |
| - Authenticates RFID Tag data received at POST '/api/authenticate'|
| - Ensure SSL certificates are in place for HTTPS.                |
| Want to contribute? https://github.com/hackpgh/dingus      |
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

	if cfg.UseAutoTLS {
		log.Fatal(autotls.Run(router, "hackpgh.org", "hackpittsburgh.com"))
	} else {
		if cfg.CertFile != "" && cfg.KeyFile != "" {
			err = router.RunTLS(":443", cfg.CertFile, cfg.KeyFile)
		} else {
			err = router.Run(":8080")
		}
		if err != nil {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}
}
