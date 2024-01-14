package services

import (
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAllRFIDs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"tag_id"}).
		AddRow(11111).
		AddRow(22222)

	mock.ExpectQuery("SELECT tag_id FROM members").WillReturnRows(rows)

	dbService := NewDBService(db, nil)

	rfids, err := dbService.GetAllRFIDs()

	assert.NoError(t, err)
	assert.Len(t, rfids, 2)
	assert.Equal(t, uint32(11111), rfids[0])
	assert.Equal(t, uint32(22222), rfids[1])

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetRFIDsForMachine(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"tag_id"}).
		AddRow(12345).
		AddRow(67890)

	// Simplified and more flexible regular expression
	mock.ExpectQuery("SELECT m.tag_id FROM members_trainings_link").
		WithArgs("MachineA").
		WillReturnRows(rows)

	dbService := NewDBService(db, nil)

	tags, err := dbService.GetRFIDsForMachine("MachineA")

	assert.NoError(t, err)
	if assert.Len(t, tags, 2) { // This check will prevent the panic by ensuring there are two elements
		assert.Equal(t, uint32(12345), tags[0])
		assert.Equal(t, uint32(67890), tags[1])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestInsertOrUpdateMembers(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	tx, err := db.Begin()
	require.NoError(t, err)

	allRFIDs := []uint32{1234, 5678}
	for _, rfid := range allRFIDs {
		mock.ExpectExec("INSERT INTO members \\(tag_id, membership_level\\) .*").
			WithArgs(rfid, 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}
	dbService := NewDBService(db, nil)

	err = dbService.insertOrUpdateMembers(tx, allRFIDs)
	assert.NoError(t, err)

	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestInsertTrainings(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	tx, err := db.Begin()
	require.NoError(t, err)

	trainingMap := map[string][]uint32{
		"Metal Lathe": {1234},
		"CNC":         {5678},
	}
	for training, rfid := range trainingMap {
		mock.ExpectExec("INSERT INTO trainings \\(training_name\\) .*").
			WithArgs(rfid, training).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	dbService := NewDBService(db, nil)

	err = dbService.insertTrainings(tx, trainingMap)
	assert.NoError(t, err)

	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestManageMemberTrainingLinks(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	tx, err := db.Begin()
	require.NoError(t, err)

	trainingMap := map[string][]uint32{
		"Metal Lathe": {1234},
		"CNC":         {5678},
	}

	for training, rfid := range trainingMap {
		mock.ExpectExec("INSERT INTO members_trainings_link .*").
			WithArgs(rfid, training).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	dbService := NewDBService(db, nil)

	err = dbService.manageMemberTrainingLinks(tx, trainingMap)
	assert.NoError(t, err)

	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDeleteInactiveMembers(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	tx, err := db.Begin()
	require.NoError(t, err)

	allRFIDs := []uint32{1234, 5678}
	tagIDList := "1234,5678"
	mock.ExpectExec(fmt.Sprintf("DELETE FROM members WHERE tag_id NOT IN \\(%s\\)", tagIDList)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	dbService := NewDBService(db, nil)

	err = dbService.deleteInactiveMembers(tx, allRFIDs)
	assert.NoError(t, err)

	err = dbService.deleteInactiveMembers(tx, allRFIDs)
	assert.NoError(t, err)

	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
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
