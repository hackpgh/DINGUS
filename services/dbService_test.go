package services

import (
	"database/sql"
	"rfid-backend/config"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockConfig() *config.Config {
	return &config.Config{
		CertFile:             "path/to/test/cert.pem",
		KeyFile:              "path/to/test/key.pem",
		DatabasePath:         "path/to/test/database.db",
		RFIDFieldName:        "RFID",
		TrainingFieldName:    "Training",
		WildApricotAccountId: 12345,
		ContactFilterQuery:   "status eq Active or status eq 'Pending - Renewal'",
	}
}

func setupTestDB(t *testing.T) *sql.DB {
	// Create an in-memory SQLite database
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create Members table
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS members (
            tag_id INTEGER PRIMARY KEY,
            membership_level INTEGER NOT NULL
        );`)
	require.NoError(t, err)

	// Create SafetyTrainings table
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS trainings (
            training_name TEXT PRIMARY KEY
        );`)
	require.NoError(t, err)

	// Create SafetyTrainingMembersLink table
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS members_trainings_link (
            tag_id INTEGER NOT NULL,
            training_name TEXT NOT NULL,
            FOREIGN KEY (tag_id) REFERENCES members(tag_id),
            FOREIGN KEY (training_name) REFERENCES trainings(training_name),
            UNIQUE (tag_id, training_name)
        );`)
	require.NoError(t, err)

	return db
}

func TestGetAllRFIDs(t *testing.T) {
	cfg := mockConfig()

	db := setupTestDB(t)
	defer db.Close()

	// Insert test data into members table
	_, err := db.Exec("INSERT INTO members (tag_id, membership_level) VALUES (11111, 1), (22222, 1)")
	require.NoError(t, err)

	dbService := NewDBService(db, cfg)

	// Execute the test function
	rfids, err := dbService.GetAllRFIDs()

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, rfids, 2)
	assert.Equal(t, uint32(11111), rfids[0])
	assert.Equal(t, uint32(22222), rfids[1])
}

func TestGetRFIDsForMachine(t *testing.T) {
	cfg := mockConfig()

	db := setupTestDB(t)
	defer db.Close()

	// Insert test data into members table
	_, err := db.Exec("INSERT INTO members (tag_id, membership_level) VALUES (12345, 1), (67890, 1)")
	require.NoError(t, err)

	// Insert test data into members_trainings_link table
	_, err = db.Exec("INSERT INTO members_trainings_link (tag_id, training_name) VALUES (12345, 'MachineA'), (67890, 'MachineA')")
	require.NoError(t, err)

	dbService := NewDBService(db, cfg)

	tags, err := dbService.GetRFIDsForMachine("MachineA")
	assert.NoError(t, err)
	assert.Len(t, tags, 2)
	assert.Equal(t, uint32(12345), tags[0])
	assert.Equal(t, uint32(67890), tags[1])
}

func TestInsertOrUpdateMembers(t *testing.T) {
	cfg := mockConfig()

	db := setupTestDB(t)

	// Start a transaction
	tx, err := db.Begin()
	require.NoError(t, err)

	dbService := NewDBService(db, cfg)

	allRFIDs := []uint32{1234, 5678}
	err = dbService.insertOrUpdateMembers(tx, allRFIDs)
	assert.NoError(t, err)

	// Commit the transaction
	err = tx.Commit()
	assert.NoError(t, err)
	// Verify the data
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM members").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestInsertTrainings(t *testing.T) {
	cfg := mockConfig()

	db := setupTestDB(t)

	tx, err := db.Begin()
	require.NoError(t, err)

	dbService := NewDBService(db, cfg)

	trainingMap := map[string][]uint32{
		"Metal Lathe": {1234},
		"CNC":         {5678},
	}
	err = dbService.insertTrainings(tx, trainingMap)
	assert.NoError(t, err)

	err = tx.Commit()
	assert.NoError(t, err)

	// Verify the data
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM trainings").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, len(trainingMap), count)
}

func TestManageMemberTrainingLinks(t *testing.T) {
	cfg := mockConfig()

	db := setupTestDB(t)

	// Start a transaction
	tx, err := db.Begin()
	require.NoError(t, err)

	dbService := NewDBService(db, cfg)

	trainingMap := map[string][]uint32{
		"Metal Lathe": {1234},
		"CNC":         {5678},
	}
	err = dbService.manageMemberTrainingLinks(tx, trainingMap)
	assert.NoError(t, err)

	err = tx.Commit()
	assert.NoError(t, err)

	// Verify the data
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM members_trainings_link").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 2, count) // Assuming each training has one RFID
}

func TestDeleteInactiveMembers(t *testing.T) {
	cfg := mockConfig()

	db := setupTestDB(t)
	defer db.Close()

	// Insert 2 test members into members table
	_, err := db.Exec("INSERT INTO members (tag_id, membership_level) VALUES (1234, 1), (67890, 1)")
	require.NoError(t, err)

	// Start a transaction
	tx, err := db.Begin()
	require.NoError(t, err)

	dbService := NewDBService(db, cfg)

	allRFIDs := []uint32{1234} // only 1 active member, remove the inactive record
	err = dbService.deleteInactiveMembers(tx, allRFIDs)
	assert.NoError(t, err)

	// Commit the transaction
	err = tx.Commit()
	assert.NoError(t, err)

	// Verify that only the active members remain
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM members").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count) // Expect only 1 active member in the table
}
