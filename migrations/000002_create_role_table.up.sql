CREATE TABLE IF NOT EXISTS role (
    id serial PRIMARY KEY,
    name citext UNIQUE NOT NULL,
    version integer NOT NULL DEFAULT 1,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);