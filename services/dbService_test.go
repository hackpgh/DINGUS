package services

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	db := setupTestDB(t)
	defer db.Close()

	// Insert test data into members table
	_, err := db.Exec("INSERT INTO members (tag_id, membership_level) VALUES (11111, 1), (22222, 1)")
	require.NoError(t, err)

	dbService := NewDBService(db, nil)

	// Execute the test function
	rfids, err := dbService.GetAllRFIDs()

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, rfids, 2)
	assert.Equal(t, uint32(11111), rfids[0])
	assert.Equal(t, uint32(22222), rfids[1])
}

func TestGetRFIDsForMachine(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test data into members table
	_, err := db.Exec("INSERT INTO members (tag_id, membership_level) VALUES (12345, 1), (67890, 1)")
	require.NoError(t, err)

	// Insert test data into members_trainings_link table
	_, err = db.Exec("INSERT INTO members_trainings_link (tag_id, training_name) VALUES (12345, 'MachineA'), (67890, 'MachineA')")
	require.NoError(t, err)

	dbService := NewDBService(db, nil)

	tags, err := dbService.GetRFIDsForMachine("MachineA")
	assert.NoError(t, err)
	assert.Len(t, tags, 2)
	assert.Equal(t, uint32(12345), tags[0])
	assert.Equal(t, uint32(67890), tags[1])
}

func TestInsertOrUpdateMembers(t *testing.T) {
	db := setupTestDB(t)

	// Start a transaction
	tx, err := db.Begin()
	require.NoError(t, err)

	dbService := NewDBService(db, nil)

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
	db := setupTestDB(t)

	tx, err := db.Begin()
	require.NoError(t, err)

	dbService := NewDBService(db, nil)

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
	db := setupTestDB(t)

	// Start a transaction
	tx, err := db.Begin()
	require.NoError(t, err)

	dbService := NewDBService(db, nil)

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
	db := setupTestDB(t)
	defer db.Close()

	// Insert 2 test members into members table
	_, err := db.Exec("INSERT INTO members (tag_id, membership_level) VALUES (1234, 1), (67890, 1)")
	require.NoError(t, err)

	// Start a transaction
	tx, err := db.Begin()
	require.NoError(t, err)

	dbService := NewDBService(db, nil)

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

// func TestProcessContactsData(t *testing.T) {
// 	db, mock, err := sqlmock.New()
// 	if err != nil {
// 		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
// 	}
// 	defer db.Close()

// 	cfg := &config.Config{
// 		RFIDFieldName:     "Door Key",
// 		TrainingFieldName: "Safety Training",
// 	}

// 	dbService := NewDBService(db, cfg)

// 	contacts := []models.Contact{
// 		{
// 			FirstName:          "testFirstName2",
// 			LastName:           "testLastname2",
// 			Email:              "test@test.com",
// 			DisplayName:        "testLastname2, testFirstName2",
// 			Organization:       "",
// 			ProfileLastUpdated: "2024-01-12T05:13:20.657-05:00",
// 			FieldValues: []models.FieldValue{
// 				{
// 					FieldName:  "Door Key",
// 					Value:      "1234",
// 					SystemCode: "Active",
// 				},
// 				{
// 					FieldName: "Safety Training",
// 					Value: []models.SafetyTraining{
// 						{
// 							Id:    20397564,
// 							Label: "Metal Lathe",
// 						},
// 						{
// 							Id:    20397565,
// 							Label: "CNC",
// 						},
// 						{
// 							Id:    20397566,
// 							Label: "Embroidery Machine",
// 						},
// 					},
// 					SystemCode: "custom-15916280",
// 				},
// 			},
// 		},
// 	}

// 	mock.ExpectBegin()

// 	// Example member ID extracted from contact data
// 	exampleMemberID := 1234 // Assuming '1234' is a member ID

// 	// Mock INSERT OR UPDATE for member
// 	insertOrUpdateMemberRegex := `INSERT INTO members.*VALUES.*ON CONFLICT.*DO UPDATE SET.*membership_level.*=.*EXCLUDED.membership_level`
// 	mock.ExpectExec(insertOrUpdateMemberRegex).
// 		WithArgs(1234, 1).
// 		WillReturnResult(sqlmock.NewResult(1, 1))

// 		// Mock INSERT for each training
// 	for _, training := range contacts[0].FieldValues[1].Value.([]models.SafetyTraining) {
// 		mock.ExpectExec("INSERT INTO trainings \\(training_name\\) VALUES \\(\\?\\) ON CONFLICT DO NOTHING").
// 			WithArgs(training.Label).
// 			WillReturnResult(sqlmock.NewResult(1, 1))

// 		// Mock INSERT for members_trainings_link
// 		mock.ExpectExec("INSERT INTO members_trainings_link \\(tag_id, training_name\\) VALUES \\(\\?, \\?\\) ON CONFLICT DO NOTHING").
// 			WithArgs(exampleMemberID, training.Label).
// 			WillReturnResult(sqlmock.NewResult(1, 1))
// 	}

// 	// Mock DELETE for stale members
// 	mock.ExpectExec("DELETE FROM members WHERE tag_id NOT IN \\(.+\\)").
// 		WillReturnResult(sqlmock.NewResult(0, 1))

// 	mock.ExpectCommit()

// 	err = dbService.ProcessContactsData(contacts)

// 	assert.NoError(t, err)

// 	if err := mock.ExpectationsWereMet(); err != nil {
// 		t.Errorf("there were unfulfilled expectations: %s", err)
// 	}
// }
