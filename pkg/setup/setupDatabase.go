// File: setup/setupDatabase.go
package setup

import (
	"database/sql"
	"rfid-backend/pkg/config"
	"rfid-backend/pkg/db"

	"github.com/sirupsen/logrus"
)

func SetupDatabase(cfg *config.Config, logger *logrus.Logger) (*sql.DB, error) {
	database, err := db.InitDB(cfg.DatabasePath)
	if err != nil {
		logger.Errorf("Failed to initialize database: %v", err)
		return nil, err
	}
	return database, nil
}
