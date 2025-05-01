CREATE TABLE IF NOT EXISTS token (
    hash bytea PRIMARY KEY,
    appuser_id cuid NOT NULL,
    expiry timestamp(0) with time zone NOT NULL,
    scope text NOT NULL,
    FOREIGN KEY (appuser_id) REFERENCES appuser(id) ON DELETE CASCADE
);