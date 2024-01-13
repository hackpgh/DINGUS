package services

import (
	"database/sql"
	"errors"
	"fmt"
	"rfid-backend/config"
	"rfid-backend/models"
	"strconv"
	"strings"
)

type DBService struct {
	db  *sql.DB
	cfg *config.Config
}

func NewDBService(db *sql.DB, cfg *config.Config) *DBService {
	return &DBService{db: db, cfg: cfg}
}

func (s *DBService) GetRFIDsForMachine(machineName string) ([]uint32, error) {
	// Fetching link table rows using the helper function
	return s.fetchRFIDs(GetRFIDsForMachineQuery, machineName)
}

func (s *DBService) GetAllRFIDs() ([]uint32, error) {
	// Fetching all RFID tags using the helper function
	return s.fetchRFIDs(GetAllRFIDsQuery)
}

// Helper function to fetch and format RFID data from a provided query.
func (s *DBService) fetchRFIDs(query string, args ...interface{}) ([]uint32, error) {
	var rfids []uint32

	// Execute the query
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate over the rows and scan the RFID values
	for rows.Next() {
		var rfid uint32
		if err := rows.Scan(&rfid); err != nil {
			return nil, err
		}
		rfids = append(rfids, rfid)
	}

	return rfids, nil
}

// service starts with this func
func (service *DBService) ProcessContactsData(contacts []models.Contact) error {
	tx, err := service.db.Begin()
	if err != nil {
		return err
	}

	// Initialize RFID slice and training map
	var allRFIDs []uint32
	trainingMap := make(map[string][]uint32)

	for _, contact := range contacts {
		rfid, trainingLabels, err := service.extractContactData(contact)
		if err != nil {
			tx.Rollback()
			return err
		}

		// Add RFID to the slice
		allRFIDs = append(allRFIDs, rfid)

		// Process training labels
		for _, label := range trainingLabels {
			trainingMap[label] = append(trainingMap[label], rfid)
		}
	}

	// Perform database operations with extracted data
	if err := service.processDatabaseUpdatesAndDeletes(tx, allRFIDs, trainingMap); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (service *DBService) extractContactData(contact models.Contact) (uint32, []string, error) {
	var rfid uint32
	var trainingLabels []string

	for _, fieldValue := range contact.FieldValues {
		switch fieldValue.FieldName {
		case service.cfg.RFIDFieldName:
			if rfidStr, ok := fieldValue.Value.(string); ok {
				rfidVal, err := strconv.ParseUint(rfidStr, 10, 32)
				if err != nil {
					return 0, nil, err
				}
				rfid = uint32(rfidVal)
			} else {
				return 0, nil, errors.New("RFID value is not a string")
			}
		case service.cfg.TrainingFieldName:
			if trainings, ok := fieldValue.Value.([]models.SafetyTraining); ok {
				for _, training := range trainings {
					trainingLabels = append(trainingLabels, training.Label)
				}
			} else {
				return 0, nil, errors.New("Training value is not a slice of SafetyTraining")
			}
		}
	}

	return rfid, trainingLabels, nil
}

func (service *DBService) processDatabaseUpdatesAndDeletes(tx *sql.Tx, allRFIDs []uint32, trainingMap map[string][]uint32) error {
	if err := service.insertOrUpdateMembers(tx, allRFIDs); err != nil {
		return err
	}

	if err := service.insertTrainings(tx, trainingMap); err != nil {
		return err
	}

	if err := service.manageMemberTrainingLinks(tx, trainingMap); err != nil {
		return err
	}

	// Prune inactive members
	if err := service.deleteInactiveMembers(tx, allRFIDs); err != nil {
		return err
	}

	return nil
}

func (service *DBService) insertOrUpdateMembers(tx *sql.Tx, allRFIDs []uint32) error {
	memberStmt, err := tx.Prepare(InsertOrUpdateMemberQuery)
	if err != nil {
		return err
	}
	defer memberStmt.Close()
	for _, rfid := range allRFIDs {
		membershipLevel := 1 // Placeholder for actual membership level
		if _, err := memberStmt.Exec(rfid, membershipLevel); err != nil {
			return err
		}
	}
	return nil
}

func (service *DBService) insertTrainings(tx *sql.Tx, trainingMap map[string][]uint32) error {
	trainingStmt, err := tx.Prepare(InsertTrainingQuery)
	if err != nil {
		return err
	}
	defer trainingStmt.Close()
	for trainingName := range trainingMap {
		if _, err := trainingStmt.Exec(trainingName); err != nil {
			return err
		}
	}
	return nil
}

func (service *DBService) manageMemberTrainingLinks(tx *sql.Tx, trainingMap map[string][]uint32) error {
	linkStmt, err := tx.Prepare(InsertMemberTrainingLinkQuery)
	if err != nil {
		return err
	}
	defer linkStmt.Close()
	for trainingName, rfids := range trainingMap {
		for _, rfid := range rfids {
			if _, err := linkStmt.Exec(rfid, trainingName); err != nil {
				return err
			}
		}
	}
	return nil
}

func (service *DBService) deleteInactiveMembers(tx *sql.Tx, allRFIDs []uint32) error {
	// Convert allRFIDs to a string slice for query
	var params []string
	for _, rfid := range allRFIDs {
		params = append(params, strconv.Itoa(int(rfid)))
	}
	tagIDList := strings.Join(params, ",")
	query := fmt.Sprintf(deleteInactiveMembersQuery, tagIDList)

	_, err := tx.Exec(query)
	return err
}
