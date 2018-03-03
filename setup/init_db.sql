-- SQLITE

DROP TABLE if EXISTS pull_request;
CREATE TABLE pull_request (number INTEGER PRIMARY KEY,latest_commit_sha1 VARCHAR(255));

ALTER TABLE pull_request ADD COLUMN updated_at TEXT;
