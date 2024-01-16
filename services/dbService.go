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
	return s.fetchRFIDs(GetRFIDsForMachineQuery, machineName)
}

func (s *DBService) GetAllRFIDs() ([]uint32, error) {
	return s.fetchRFIDs(GetAllRFIDsQuery)
}

// Helper function to fetch and format RFID data from a provided query.
func (s *DBService) fetchRFIDs(query string, args ...interface{}) ([]uint32, error) {
	var rfids []uint32

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
	var allRFIDs []uint32
	trainingMap := make(map[string][]uint32)

	for _, contact := range contacts {
		rfid, trainingLabels, err := s.extractContactData(contact)
		if err != nil {
			return err
		}

		if rfid != 0 {
			allRFIDs = append(allRFIDs, rfid)
		}

		for _, label := range trainingLabels {
			trainingMap[label] = append(trainingMap[label], rfid)
		}
	}

	// Guard against empty WA contacts responses which is the
	// typical first response from WA API when WA async
	// resultId is refreshing
	if len(allRFIDs) <= 0 {
		return errors.New("allRFIDs list, parsed from Wild Apricot, was empty")
	}

	totalContactsMissingRFIDs := len(contacts) - len(allRFIDs)
	if totalContactsMissingRFIDs > 0 {
		log.Printf("Total empty RFID values detected: %d", totalContactsMissingRFIDs)
		log.Println("Will ignore contact if awaiting onboarding, otherwise deleting member")
	}

	// Perform database operations with extracted data
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	if err := s.processDatabaseUpdatesAndDeletes(tx, allRFIDs, trainingMap); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *DBService) extractContactData(contact models.Contact) (uint32, []string, error) {
	var convertedRfid uint32
	var trainingLabels []string

	for _, fieldValue := range contact.FieldValues {
		switch fieldValue.FieldName {
		case s.cfg.RFIDFieldName:
			rfid, err := s.parseRFID(fieldValue)
			if err != nil {
				return 0, nil, fmt.Errorf("error parsing RFID for contact %d: %v", contact.Id, err)
			}
			convertedRfid = rfid

		case s.cfg.TrainingFieldName:
			labels, err := s.parseTrainingLabels(fieldValue)
			if err != nil {
				return 0, nil, fmt.Errorf("error parsing training labels for contact %d: %v", contact.Id, err)
			}
			trainingLabels = append(trainingLabels, labels...)
		}
	}

	return convertedRfid, trainingLabels, nil
}

func (s *DBService) parseRFID(fieldValue models.FieldValue) (uint32, error) {
	rfidValue, ok := fieldValue.Value.(string)
	if !ok {
		return 0, errors.New("RFID value is not a string")
	}

	if len(rfidValue) <= 0 {
		// Suppress error on empty RFID field value, return 0
		return uint32(0), nil
	}

	rfid, err := strconv.ParseInt(rfidValue, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to convert string RFID to int: %v", err)
	}

	if rfid <= 0 {
		return 0, errors.New("RFID value is non-positive")
	}

	return uint32(rfid), nil
}

func (s *DBService) parseTrainingLabels(fieldValue models.FieldValue) ([]string, error) {
	trainingValues, ok := fieldValue.Value.([]interface{})
	if !ok {
		return nil, errors.New("training value is not a slice")
	}

	var labels []string
	for _, t := range trainingValues {
		trainingMap, ok := t.(map[string]interface{})
		if !ok {
			return nil, errors.New("training item is not a map")
		}

		label, err := s.extractLabelFromTrainingMap(trainingMap)
		if err != nil {
			return nil, err
		}
		labels = append(labels, label)
	}

	return labels, nil
}

func (s *DBService) extractLabelFromTrainingMap(trainingMap map[string]interface{}) (string, error) {
	label, ok := trainingMap["Label"].(string)
	if !ok {
		return "", errors.New("training label is not a string")
	}

	return label, nil
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

	if err := s.deleteInactiveMembers(tx, allRFIDs); err != nil {
		return err
	}

	// Database transactions successful, no errors
	return nil
}

func (s *DBService) insertOrUpdateMembers(tx *sql.Tx, allRFIDs []uint32) error {
	memberStmt, err := tx.Prepare(InsertOrUpdateMemberQuery)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		return err
	}
	defer func() {
		memberStmt.Close()
	}()

	for _, rfid := range allRFIDs {
		membershipLevel := 1 // Placeholder for actual membership level
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
