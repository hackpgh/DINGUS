// File: setup/setupBackgroundSync.go

package setup

import (
	"rfid-backend/services"
	"time"

	"github.com/sirupsen/logrus"
)

func StartBackgroundDatabaseUpdate(waService *services.WildApricotService, dbService *services.DBService, logger *logrus.Logger) {
	go func() {
		updateEntireDatabaseFromWildApricot(waService, dbService, logger)
		ticker := time.NewTicker(30 * time.Minute)
		for range ticker.C {
			updateEntireDatabaseFromWildApricot(waService, dbService, logger)
		}
	}()
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
