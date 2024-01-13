package services

const (
	// GetRFIDsForMachineQuery retrieves RFID tags for a specific machine
	GetRFIDsForMachineQuery = `
        SELECT m.tag_id
        FROM members_trainings_link mtl
        JOIN trainings t ON mtl.training_name = t.training_name
        WHERE t.training_name = ?;
    `

	// GetAllRFIDsQuery retrieves all RFID tags
	GetAllRFIDsQuery = `
        SELECT tag_id
        FROM members;
    `

	// SQL statement to insert or update a member
	InsertOrUpdateMemberQuery = `
        INSERT INTO members (tag_id, membership_level)
        VALUES (?, ?)
        ON CONFLICT(tag_id) DO UPDATE SET membership_level = EXCLUDED.membership_level
    `

	// SQL statement to insert a training name
	InsertTrainingQuery = `
        INSERT INTO trainings (training_name)
        VALUES (?)
        ON CONFLICT DO NOTHING
    `

	// SQL statement to insert into members_trainings_link
	InsertMemberTrainingLinkQuery = `
        INSERT INTO members_trainings_link (tag_id, training_name)
        VALUES (?, ?)
        ON CONFLICT DO NOTHING
    `

	// SQL statement to delete stale members
	deleteInactiveMembersQuery = `
        DELETE FROM members WHERE tag_id NOT IN (%s)
    `

	// TODO: Should we add more queries and dbService logic to
	//       handle pruning on 'trainings' and 'members_trainings_link' tables
)
