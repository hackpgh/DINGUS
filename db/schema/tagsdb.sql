CREATE TABLE IF NOT EXISTS members (
    contact_id INTEGER PRIMARY KEY,
    tag_id INTEGER NOT NULL,
    membership_level INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_members_tag_id ON members(tag_id);


CREATE TABLE IF NOT EXISTS devices (
    ip_address TEXT NOT NULL UNIQUE,
    mac_address TEXT NOT NULL UNIQUE,
    requires_training INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS trainings (
    label TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS members_trainings_link (
    tag_id INTEGER NOT NULL,
    label TEXT NOT NULL,
    FOREIGN KEY (tag_id) REFERENCES members(tag_id),
    FOREIGN KEY (label) REFERENCES trainings(label),
    UNIQUE (tag_id, label)
);

CREATE TABLE IF NOT EXISTS devices_trainings_link (
    mac_address TEXT NOT NULL,
    label TEXT NOT NULL,
    FOREIGN KEY (label) REFERENCES trainings(label),
    FOREIGN KEY (mac_address) REFERENCES devices(mac_address),
    PRIMARY KEY (label, mac_address)
);