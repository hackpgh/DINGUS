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
	logger.WithFields(logrus.Fields{
		"action": "FetchingContacts",
		"source": "WildApricotAPI",
	}).Info("Updating database with Wild Apricot contacts")

	contacts, err := waService.GetContacts()
	if err != nil {
		logger.WithFields(logrus.Fields{
			"action": "FetchContacts",
			"status": "Failed",
			"error":  err,
		}).Error("Failed to fetch contacts from Wild Apricot")
		return
	}

	if len(contacts) <= 0 {
		logger.WithFields(logrus.Fields{
			"action": "ProcessContacts",
			"status": "NoContacts",
		}).Info("No new contacts to process from Wild Apricot")
		return
	}

	if err = dbService.ProcessContactsData(contacts); err != nil {
		logger.WithFields(logrus.Fields{
			"action": "UpdateDatabase",
			"status": "Failed",
			"error":  err,
		}).Error("Failed to update database with new contacts")
	} else {
		logger.WithFields(logrus.Fields{
			"action":            "UpdateDatabase",
			"status":            "Success",
			"contactsProcessed": len(contacts),
		}).Info("Database successfully updated with Wild Apricot contacts")
	}
}
