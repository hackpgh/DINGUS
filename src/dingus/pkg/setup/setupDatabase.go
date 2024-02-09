// File: setup/setupDatabase.go
package setup

import (
	"database/sql"
	"dingus/pkg/config"
	"dingus/pkg/db"

	"github.com/sirupsen/logrus"
)

func SetupDatabase(cfg *config.Config, logger *logrus.Logger) (*sql.DB, error) {
	logger.Infof("Database Path from config: %s", cfg.DatabasePath)
	database, err := db.InitDB(cfg.DatabasePath)
	if err != nil {
		logger.Errorf("Failed to initialize database: %v", err)
		return nil, err
	}
	return database, nil
}
