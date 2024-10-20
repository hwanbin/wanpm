CREATE TABLE IF NOT EXISTS token (
    hash bytea PRIMARY KEY,
    appuser_internal_id integer NOT NULL,
    expiry timestamp(0) with time zone NOT NULL,
    scope text NOT NULL,
    FOREIGN KEY (appuser_internal_id) REFERENCES appuser(internal_id) ON DELETE CASCADE
);