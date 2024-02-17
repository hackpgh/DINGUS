package services

const (
	TagExistsQuery = `
		SELECT EXISTS(SELECT 1 FROM members WHERE tag_id = ?)
	`

	GetTagIdsForTrainingQuery = `
        SELECT tag_id
        FROM members_trainings_link
        WHERE label = ?;
    `

	GetTrainingQuery = `
		SELECT label
		FROM trainings
		WHERE label = ?
	`

	GetAllTrainingsQuery = `
		SELECT label
		FROM trainings
	`

	GetAllDevicesQuery = `
		SELECT ip_address, mac_address
		FROM devices;
	`

	GetAllTagIdsQuery = `
        SELECT tag_id
        FROM members;
    `

	GetAllDevicesTrainingsQuery = `
	SELECT mac_address, label
	FROM devices_trainings_link;
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
        INSERT OR IGNORE INTO devices (ip_address, mac_address, requires_training)
        VALUES (?, ?, ?);
    `

	InsertDeviceTrainingLinkQuery = `
		INSERT INTO devices_trainings_link (mac_address, label)
		VALUES (?, ?)
		ON CONFLICT(mac_address, label) DO UPDATE 
		SET mac_address = EXCLUDED.mac_address;
	`

	DeleteInactiveMembersQuery = `
        DELETE FROM members WHERE contact_id NOT IN (%s)
    `

	DeleteLapsedMembersQuery = `
        DELETE FROM members WHERE contact_id = %s
    `

	DeleteDeviceTrainingLinkQuery = `
		DELETE FROM devices_trainings_link
		WHERE mac_address = ?;
	`
)
