package services

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
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
func (s *DBService) ProcessContactsData(contacts []models.Contact) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	// Initialize RFID slice and training map
	var allRFIDs []uint32
	trainingMap := make(map[string][]uint32)

	for _, contact := range contacts {
		rfid, trainingLabels, err := s.extractContactData(contact)
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
	if err := s.processDatabaseUpdatesAndDeletes(tx, allRFIDs, trainingMap); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s *DBService) extractContactData(contact models.Contact) (uint32, []string, error) {
	var converted_rfid uint32
	var trainingLabels []string

	for _, fieldValue := range contact.FieldValues {
		switch fieldValue.FieldName {
		case s.cfg.RFIDFieldName:
			rfid, err := strconv.ParseInt(fieldValue.Value.(string), 10, 32)
			if err != nil {
				log.Print("Failed to convert string rfid to int")
			}

			if rfid > 0 {
				converted_rfid = uint32(rfid)
			} else {
				return 0, nil, errors.New("RFID value is not an int")
			}
		case s.cfg.TrainingFieldName:
			trainingValues, ok := fieldValue.Value.([]interface{})
			if !ok {
				return 0, nil, errors.New("Training value is not a slice")
			}

			for _, t := range trainingValues {
				trainingMap, ok := t.(map[string]interface{})
				if !ok {
					return 0, nil, errors.New("Training item is not a map")
				}

				// Assuming 'Id' and 'Label' are the keys in the map
				trainingId, ok := trainingMap["Id"].(float64)
				if !ok {
					return 0, nil, errors.New("Training Id is not a float64")
				}

				trainingLabel, ok := trainingMap["Label"].(string)
				if !ok {
					return 0, nil, errors.New("Training Label is not a string")
				}

				training := models.SafetyTraining{
					Id:    int(trainingId),
					Label: trainingLabel,
				}

				trainingLabels = append(trainingLabels, training.Label)
			}
		}
	}

	return converted_rfid, trainingLabels, nil
}

func (s *DBService) processDatabaseUpdatesAndDeletes(tx *sql.Tx, allRFIDs []uint32, trainingMap map[string][]uint32) error {
	if err := s.insertOrUpdateMembers(tx, allRFIDs); err != nil {
		return err
	}

	if err := s.insertTrainings(tx, trainingMap); err != nil {
		return err
	}

	if err := s.manageMemberTrainingLinks(tx, trainingMap); err != nil {
		return err
	}

	// Prune inactive members
	if err := s.deleteInactiveMembers(tx, allRFIDs); err != nil {
		return err
	}

	return nil
}

func (s *DBService) insertOrUpdateMembers(tx *sql.Tx, allRFIDs []uint32) error {
	memberStmt, err := tx.Prepare(InsertOrUpdateMemberQuery)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		return err
	}
	defer func() {
		log.Println("Closing memberStmt")
		memberStmt.Close()
	}()

	for _, rfid := range allRFIDs {
		membershipLevel := 1 // Placeholder for actual membership level
		log.Printf("Executing insertOrUpdate for RFID: %d", rfid)
		if _, err := memberStmt.Exec(rfid, membershipLevel); err != nil {
			log.Printf("Error executing insertOrUpdate for RFID %d: %v", rfid, err)
			return err
		}
	}
	return nil
}

func (s *DBService) insertTrainings(tx *sql.Tx, trainingMap map[string][]uint32) error {
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

func (s *DBService) manageMemberTrainingLinks(tx *sql.Tx, trainingMap map[string][]uint32) error {
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

func (s *DBService) deleteInactiveMembers(tx *sql.Tx, allRFIDs []uint32) error {
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
