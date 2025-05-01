CREATE TABLE IF NOT EXISTS appuser (
    id cuid PRIMARY KEY,
    email citext UNIQUE NOT NULL,
    first_name citext NOT NULL,
    last_name citext NOT NULL,
    password_hash bytea NOT NULL,
    activated bool NOT NULL,
    version integer NOT NULL DEFAULT 1,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);