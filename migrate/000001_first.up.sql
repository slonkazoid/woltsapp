CREATE TABLE groups (
	id TEXT UNIQUE NOT NULL
);

CREATE TABLE defined_hosts (
	name TEXT UNIQUE NOT NULL,
	mac_address TEXT NOT NULL
);
