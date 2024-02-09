package services

const (
	GetTagIdsForTrainingQuery = `
        SELECT tag_id
        FROM members_trainings_link
        WHERE label = $1;
    `

	GetTrainingQuery = `
        SELECT label
        FROM trainings
        WHERE label = $1
    `

	GetAllTrainingsQuery = `
        SELECT label
        FROM trainings
    `

	GetAllDevicesQuery = `
        SELECT ip_address
        FROM devices;
    `

	GetAllTagIdsQuery = `
        SELECT tag_id
        FROM members;
    `

	InsertOrUpdateMemberQuery = `
        INSERT INTO members (contact_id, tag_id, membership_level)
        VALUES ($1, $2, $3)
        ON CONFLICT(contact_id) DO UPDATE SET tag_id = EXCLUDED.tag_id, membership_level = EXCLUDED.membership_level;
    `

	InsertTrainingQuery = `
        INSERT INTO trainings (label)
        VALUES ($1)
        ON CONFLICT (label) DO NOTHING;
    `

	InsertMemberTrainingLinkQuery = `
        INSERT INTO members_trainings_link (tag_id, label)
        VALUES ($1, $2)
        ON CONFLICT (tag_id, label) DO NOTHING;
    `

	InsertDeviceQuery = `
        INSERT INTO devices (ip_address, requires_training)
        VALUES ($1, $2)
        ON CONFLICT (ip_address) DO NOTHING;
    `

	InsertDeviceTrainingLinkQuery = `
        INSERT INTO devices_trainings_link (ip_address, label)
        VALUES ($1, $2)
        ON CONFLICT (ip_address, label) DO NOTHING;
    `

	// For dynamic IN clause handling in PostgreSQL, you will likely need to use a custom function or prepare the query dynamically in your application code.
	DeleteInactiveMembersQuery = `
        DELETE FROM members WHERE contact_id NOT IN (%s);
    `

	// Deleting lapsed members can stay the same with parameter substitution handled in your application code.
	DeleteLapsedMembersQuery = `
        DELETE FROM members WHERE contact_id = $1
    `

	DeleteDeviceTrainingLinkQuery = `
        DELETE FROM devices_trainings_link
        WHERE ip_address = $1;
    `
)
