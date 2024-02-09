CREATE TABLE IF NOT EXISTS members (
    contact_id INTEGER PRIMARY KEY,
    tag_id INTEGER NOT NULL UNIQUE,
    membership_level INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_members_tag_id ON members(tag_id);

CREATE TABLE IF NOT EXISTS devices (
    ip_address TEXT NOT NULL UNIQUE,
    requires_training BOOLEAN NOT NULL -- Assuming this is a true/false value
);

CREATE TABLE IF NOT EXISTS trainings (
    label TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS members_trainings_link (
    tag_id INTEGER NOT NULL,
    label TEXT NOT NULL,
    FOREIGN KEY (tag_id) REFERENCES members(tag_id) ON DELETE CASCADE,
    FOREIGN KEY (label) REFERENCES trainings(label) ON DELETE CASCADE,
    UNIQUE (tag_id, label)
);

CREATE TABLE IF NOT EXISTS devices_trainings_link (
    ip_address TEXT NOT NULL,
    label TEXT NOT NULL,
    FOREIGN KEY (label) REFERENCES trainings(label) ON DELETE CASCADE,
    FOREIGN KEY (ip_address) REFERENCES devices(ip_address) ON DELETE CASCADE,
    PRIMARY KEY (label, ip_address)
);