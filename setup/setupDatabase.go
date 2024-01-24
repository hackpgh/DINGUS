package setup

import (
	"database/sql"
	"rfid-backend/config"
	"rfid-backend/db"

	"github.com/sirupsen/logrus"
)

func setupDatabase(cfg *config.Config, logger *logrus.Logger) (*sql.DB, error) {
	database, err := db.InitDB(cfg.DatabasePath)
	if err != nil {
		logger.Errorf("Failed to initialize database: %v", err)
		return nil, err
	}
	return database, nil
}
