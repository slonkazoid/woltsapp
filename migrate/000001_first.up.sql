CREATE TABLE allowed_users (
	phone_no TEXT UNIQUE NOT NULL,
	permission_level INTEGER NOT NULL
);

CREATE TABLE defined_hosts (
	name TEXT UNIQUE NOT NULL,
	mac_address TEXT NOT NULL
);
