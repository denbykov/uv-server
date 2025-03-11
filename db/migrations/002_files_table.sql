CREATE TABLE file_statuses (
	status TEXT PRIMARY KEY,
	description TEXT
);

INSERT INTO file_statuses (status, description)
VALUES 
	('p', 'Pending'),
	('d', 'Downloading'),
	('f', 'Finished');

CREATE TABLE sources (
	source TEXT PRIMARY KEY,
	description TEXT
);

INSERT INTO sources (source, description)
VALUES 
	('yt', 'Youtube');

CREATE TABLE files (
	id INTEGER PRIMARY KEY,
	path TEXT NULL UNIQUE,
	source_url TEXT NOT NULL UNIQUE,
	source TEXT NOT NULL,
	status TEXT NOT NULL,
	added_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY(status) REFERENCES file_statuses(status)
	FOREIGN KEY(source) REFERENCES sources(source)
);
