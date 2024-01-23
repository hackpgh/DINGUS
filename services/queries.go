package services

const (
	// GetTagIdsForMachineQuery retrieves TagId tags for a specific machine
	GetTagIdsForMachineQuery = `
        SELECT tag_id
        FROM members_trainings_link
        WHERE training_name = ?;
    `

	// GetAllTagIdsQuery retrieves all TagId tags
	GetAllTagIdsQuery = `
        SELECT tag_id
        FROM members;
    `

	// SQL statement to insert or update a member
	InsertOrUpdateMemberQuery = `
		INSERT OR IGNORE INTO members (contact_id, tag_id, membership_level)
		VALUES (?, ?, ?)
		ON CONFLICT(contact_id) DO UPDATE SET tag_id = ?, membership_level = EXCLUDED.membership_level;
	`

	// SQL statement to insert a training name
	InsertTrainingQuery = `
        INSERT OR IGNORE INTO trainings (training_name)
        VALUES (?);
    `

	// SQL statement to insert into members_trainings_link
	InsertMemberTrainingLinkQuery = `
        INSERT OR IGNORE INTO members_trainings_link (tag_id, training_name)
        VALUES (?, ?);
    `

	// POSTGRES Compatible queries
	// // SQL statement to insert or update a member
	// InsertOrUpdateMemberQuery = `
	//     INSERT INTO members (tag_id, membership_level)
	//     VALUES (?, ?)
	//     ON CONFLICT(tag_id) DO UPDATE SET membership_level = EXCLUDED.membership_level
	// `

	// // SQL statement to insert a training name
	// InsertTrainingQuery = `
	//     INSERT INTO trainings (training_name)
	//     VALUES (?)
	//     ON CONFLICT DO NOTHING
	// `

	// // SQL statement to insert into members_trainings_link
	// InsertMemberTrainingLinkQuery = `
	//     INSERT INTO members_trainings_link (tag_id, training_name)
	//     VALUES (?, ?)
	//     ON CONFLICT DO NOTHING
	// `

	// SQL statement to delete stale members
	deleteInactiveMembersQuery = `
        DELETE FROM members WHERE contact_id NOT IN (%s)
    `

	deleteLapsedMembersQuery = `
        DELETE FROM members WHERE contact_id = %s
    `

	// TODO: Should we add more queries and dbService logic to
	//       handle pruning on 'trainings' and 'members_trainings_link' tables
)
