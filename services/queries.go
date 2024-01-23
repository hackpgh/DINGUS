package services

const (
	GetTagIdsForTrainingQuery = `
        SELECT tag_id
        FROM members_trainings_link
        WHERE label = ?;
    `

	GetTraining = `
		SELECT label
		FROM trainings
		WHERE label = ?
	`

	GetAllTagIdsQuery = `
        SELECT tag_id
        FROM members;
    `

	InsertOrUpdateMemberQuery = `
		INSERT OR IGNORE INTO members (contact_id, tag_id, membership_level)
		VALUES (?, ?, ?)
		ON CONFLICT(contact_id) DO UPDATE SET tag_id = ?, membership_level = EXCLUDED.membership_level;
	`

	InsertTrainingQuery = `
        INSERT OR IGNORE INTO trainings (label)
        VALUES (?);
    `

	InsertMemberTrainingLinkQuery = `
        INSERT OR IGNORE INTO members_trainings_link (tag_id, label)
        VALUES (?, ?);
    `

	InsertDeviceQuery = `
        INSERT OR IGNORE INTO devices (ip_address)
        VALUES (?);
    `

	InsertDeviceTrainingLinkQuery = `
		INSERT OR IGNORE INTO devices_trainings_link (ip_address, label)
		VALUES (?, ?)
		ON CONFLICT(label) DO UPDATE SET ip_address = ?;
	`

	deleteInactiveMembersQuery = `
        DELETE FROM members WHERE contact_id NOT IN (%s)
    `

	deleteLapsedMembersQuery = `
        DELETE FROM members WHERE contact_id = %s
    `
)
