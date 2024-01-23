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
	var tag_ids []uint32

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate over the rows and scan the TagId values
	for rows.Next() {
		var tag_id uint32
		if err := rows.Scan(&tag_id); err != nil {
			return nil, err
		}
		tag_ids = append(tag_ids, tag_id)
	}

	return tag_ids, nil
}

// service starts with this func
func (s *DBService) ProcessContactsData(contacts []models.Contact) error {
	var all_contacts []int
	var all_tag_ids []uint32
	trainingMap := make(map[string][]uint32)

	for _, contact := range contacts {
		contact_id, tag_id, trainingLabels, err := contact.ExtractContactData(s.cfg)
		if err != nil {
			return err
		}

		if contact_id != 0 {
			all_contacts = append(all_contacts, contact_id)
		}

		if tag_id != 0 {
			all_tag_ids = append(all_tag_ids, tag_id)
		}

		for _, label := range trainingLabels {
			trainingMap[label] = append(trainingMap[label], tag_id)
		}
	}

	// Guard against empty WA contacts responses which is the
	// typical first response from WA API when WA async
	// resultId is refreshing
	if len(all_tag_ids) <= 0 {
		return errors.New("all_tag_ids list, parsed from Wild Apricot, was empty")
	}

	missing_tags := len(contacts) - len(all_tag_ids)
	if missing_tags > 0 {
		log.Printf("Total empty TagId values detected: %d", missing_tags)
		log.Println("Will ignore contact if awaiting onboarding, otherwise deleting member")
	}

	// Perform database operations with extracted data
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	if err := s.processDatabaseUpdatesAndDeletes(tx, all_contacts, all_tag_ids, trainingMap); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *DBService) processDatabaseUpdatesAndDeletes(tx *sql.Tx, all_contacts []int, all_tag_ids []uint32, trainingMap map[string][]uint32) error {
	if err := s.insertOrUpdateAllMembers(tx, all_contacts, all_tag_ids); err != nil {
		return err
	}

	if err := s.insertTrainings(tx, trainingMap); err != nil {
		return err
	}

	if err := s.manageMemberTrainingLinks(tx, trainingMap); err != nil {
		return err
	}

	if err := s.deleteInactiveMembers(tx, all_contacts); err != nil {
		return err
	}

	// Database transactions successful, no errors
	return nil
}

func (s *DBService) insertOrUpdateAllMembers(tx *sql.Tx, all_contacts []int, all_tag_ids []uint32) error {
	memberStmt, err := tx.Prepare(InsertOrUpdateMemberQuery)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		return err
	}
	defer memberStmt.Close()

	for i := 0; i < len(all_contacts); i++ {
		membership_level := 1 // Placeholder for actual membership level
		if _, err := memberStmt.Exec(all_contacts[i], all_tag_ids[i], membership_level, all_tag_ids[i]); err != nil {
			log.Printf("Error executing insertOrUpdate for tag_id %d: %v", all_tag_ids[i], err)
			return err
		}
	}
	log.Printf("finished %d inserts into members table", len(all_contacts))

	return nil
}

func (s *DBService) insertActiveMember(tx *sql.Tx, contact_id int, tag_id uint32) error {
	memberStmt, err := tx.Prepare(InsertOrUpdateMemberQuery)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		return err
	}
	defer memberStmt.Close()

	membership_level := 1 // Placeholder for actual membership level
	log.Printf("contact_id: %d, tag_id: %d, ml: %d, tag_id: %d", contact_id, tag_id, membership_level, tag_id)
	if _, err := memberStmt.Exec(contact_id, tag_id, membership_level, tag_id); err != nil {
		log.Printf("Error executing insertOrUpdate for tag_id %d: %v", tag_id, err)
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
	for training_label := range trainingMap {
		if _, err := trainingStmt.Exec(training_label); err != nil {
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

	for training_label, tag_ids := range trainingMap {
		for _, tag_id := range tag_ids {
			if _, err := linkStmt.Exec(tag_id, training_label); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *DBService) deleteInactiveMembers(tx *sql.Tx, all_contacts []int) error {
	// Convert all_contacts to a string slice for query
	var params []string
	for _, contact_id := range all_contacts {
		params = append(params, strconv.Itoa(contact_id))
	}
	all_contact_ids := strings.Join(params, ",")
	query := fmt.Sprintf(deleteInactiveMembersQuery, all_contact_ids)

	_, err := tx.Exec(query)
	return err
}

func (s *DBService) deleteLapsedMember(tx *sql.Tx, contact_id int) error {
	query := fmt.Sprintf(deleteLapsedMembersQuery, strconv.Itoa(int(contact_id)))

	_, err := tx.Exec(query)
	return err
}

func (s *DBService) insertTrainingsLink(tx *sql.Tx, tag_id uint32, trainings []string) error {
	linkStmt, err := tx.Prepare(InsertMemberTrainingLinkQuery)
	if err != nil {
		return err
	}
	defer linkStmt.Close()
	for _, training_label := range trainings {
		if _, err := linkStmt.Exec(tag_id, training_label); err != nil {
			return err
		}
	}
	return nil
}

func (s *DBService) ProcessContactWebhookTrainingData(contact models.Contact) error {
	_, tag_id, training_labels, err := contact.ExtractContactData(s.cfg)
	if err != nil {
		return err
	}

	// Perform database operations with extracted data
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	log.Printf("training_labels: %+v", training_labels)
	if err := s.insertTrainingsLink(tx, tag_id, training_labels); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s *DBService) ProcessMembershipWebhook(params webhooks.MembershipParameters, contact models.Contact) error {
	contact_id, tag_id, _, err := contact.ExtractContactData(s.cfg)
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

		if err := s.deleteLapsedMember(tx, contact_id); err != nil {
			tx.Rollback()
			return err
		}
	case webhooks.StatusActive:
		log.Printf("Active membership detected")
		if err := s.insertActiveMember(tx, contact_id, tag_id); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
