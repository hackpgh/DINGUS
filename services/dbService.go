package services

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"rfid-backend/config"
	"rfid-backend/models"
	"rfid-backend/webhooks"
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

func (s *DBService) GetTagIdsForMachine(machineName string) ([]uint32, error) {
	return s.fetchTagIds(GetTagIdsForMachineQuery, machineName)
}

func (s *DBService) GetAllTagIds() ([]uint32, error) {
	return s.fetchTagIds(GetAllTagIdsQuery)
}

// Helper function to fetch and format TagId data from a provided query.
func (s *DBService) fetchTagIds(query string, args ...interface{}) ([]uint32, error) {
	var tagIds []uint32

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate over the rows and scan the TagId values
	for rows.Next() {
		var tagId uint32
		if err := rows.Scan(&tagId); err != nil {
			return nil, err
		}
		tagIds = append(tagIds, tagId)
	}

	return tagIds, nil
}

// service starts with this func
func (s *DBService) ProcessContactsData(contacts []models.Contact) error {
	var allContacts []int
	var allTagIds []uint32
	trainingMap := make(map[string][]uint32)

	for _, contact := range contacts {
		contactId, tagId, trainingLabels, err := contact.ExtractContactData(s.cfg)
		if err != nil {
			return err
		}

		if contactId != 0 {
			allContacts = append(allContacts, contactId)
		}

		if tagId != 0 {
			allTagIds = append(allTagIds, tagId)
		}

		for _, label := range trainingLabels {
			trainingMap[label] = append(trainingMap[label], tagId)
		}
	}

	// Guard against empty WA contacts responses which is the
	// typical first response from WA API when WA async
	// resultId is refreshing
	if len(allTagIds) <= 0 {
		return errors.New("allTagIds list, parsed from Wild Apricot, was empty")
	}

	missingTags := len(contacts) - len(allTagIds)
	if missingTags > 0 {
		log.Printf("Total empty TagId values detected: %d", missingTags)
		log.Println("Will ignore contact if awaiting onboarding, otherwise deleting member")
	}

	// Perform database operations with extracted data
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	if err := s.processDatabaseUpdatesAndDeletes(tx, allContacts, allTagIds, trainingMap); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *DBService) processDatabaseUpdatesAndDeletes(tx *sql.Tx, allContacts []int, allTagIds []uint32, trainingMap map[string][]uint32) error {
	if err := s.insertOrUpdateAllMembers(tx, allContacts, allTagIds); err != nil {
		return err
	}

	if err := s.insertTrainings(tx, trainingMap); err != nil {
		return err
	}

	if err := s.manageMemberTrainingLinks(tx, trainingMap); err != nil {
		return err
	}

	if err := s.deleteInactiveMembers(tx, allContacts); err != nil {
		return err
	}

	// Database transactions successful, no errors
	return nil
}

func (s *DBService) insertOrUpdateAllMembers(tx *sql.Tx, allContacts []int, allTagIds []uint32) error {
	memberStmt, err := tx.Prepare(InsertOrUpdateMemberQuery)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		return err
	}
	defer memberStmt.Close()

	for i := 0; i < len(allContacts); i++ {
		membershipLevel := 1 // Placeholder for actual membership level
		if _, err := memberStmt.Exec(allContacts[i], allTagIds[i], membershipLevel, allTagIds[i]); err != nil {
			log.Printf("Error executing insertOrUpdate for tagId %d: %v", allTagIds[i], err)
			return err
		}
	}
	log.Printf("finished %d inserts into members table", len(allContacts))

	return nil
}

func (s *DBService) insertActiveMember(tx *sql.Tx, contactId int, tagId uint32) error {
	memberStmt, err := tx.Prepare(InsertOrUpdateMemberQuery)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		return err
	}
	defer memberStmt.Close()

	membershipLevel := 1 // Placeholder for actual membership level
	log.Printf("contactId: %d, tagId: %d, ml: %d, tagId: %d", contactId, tagId, membershipLevel, tagId)
	if _, err := memberStmt.Exec(contactId, tagId, membershipLevel, tagId); err != nil {
		log.Printf("Error executing insertOrUpdate for tagId %d: %v", tagId, err)
		return err
	}
	return nil
}

func (s *DBService) insertTrainings(tx *sql.Tx, trainingMap map[string][]uint32) error {
	trainingStmt, err := tx.Prepare(InsertTrainingQuery)
	if err != nil {
		return err
	}
	defer trainingStmt.Close()
	for trainingLabel := range trainingMap {
		if _, err := trainingStmt.Exec(trainingLabel); err != nil {
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

	for trainingLabel, tagIds := range trainingMap {
		for _, tagId := range tagIds {
			if _, err := linkStmt.Exec(tagId, trainingLabel); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *DBService) deleteInactiveMembers(tx *sql.Tx, allContacts []int) error {
	// Convert allContacts to a string slice for query
	var params []string
	for _, contactId := range allContacts {
		params = append(params, strconv.Itoa(contactId))
	}
	all_contactIds := strings.Join(params, ",")
	query := fmt.Sprintf(deleteInactiveMembersQuery, all_contactIds)

	_, err := tx.Exec(query)
	return err
}

func (s *DBService) deleteLapsedMember(tx *sql.Tx, contactId int) error {
	query := fmt.Sprintf(deleteLapsedMembersQuery, strconv.Itoa(int(contactId)))

	_, err := tx.Exec(query)
	return err
}

func (s *DBService) insertTrainingsLink(tx *sql.Tx, tagId uint32, trainings []string) error {
	linkStmt, err := tx.Prepare(InsertMemberTrainingLinkQuery)
	if err != nil {
		return err
	}
	defer linkStmt.Close()
	for _, trainingLabel := range trainings {
		if _, err := linkStmt.Exec(tagId, trainingLabel); err != nil {
			return err
		}
	}
	return nil
}

func (s *DBService) ProcessContactWebhookTrainingData(params webhooks.ContactParameters, contact models.Contact) error {
	contactId, tagId, trainingLabels, err := contact.ExtractContactData(s.cfg)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	// Handle tagId changes
	// Delete from the members table if they do not have a valid tagId
	if tagId <= 0 {
		if err := s.deleteLapsedMember(tx, contactId); err != nil {
			return err
		}

		return tx.Commit()
	}

	// If and only if Status is active, attempt to insert the active member
	if contact.Status == "Active" {
		if err := s.insertActiveMember(tx, contactId, tagId); err != nil {
			return err
		}
	}

	// Handle trainings changes
	log.Printf("trainingLabels: %+v", trainingLabels)
	if err := s.insertTrainingsLink(tx, tagId, trainingLabels); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s *DBService) ProcessMembershipWebhook(params webhooks.MembershipParameters, contact models.Contact) error {
	contactId, tagId, _, err := contact.ExtractContactData(s.cfg)
	if err != nil {
		return err
	}

	// Perform database operations with extracted data
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	switch params.MembershipStatus {
	case webhooks.StatusLapsed:
		log.Printf("Lapsed membership detected")

		if err := s.deleteLapsedMember(tx, contactId); err != nil {
			tx.Rollback()
			return err
		}
	case webhooks.StatusActive:
		log.Printf("Active membership detected")
		if err := s.insertActiveMember(tx, contactId, tagId); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
