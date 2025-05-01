CREATE TABLE IF NOT EXISTS activity (
    id serial PRIMARY KEY,
    name citext UNIQUE NOT NULL,
    version integer NOT NULL DEFAULT 1,
    created_at timestamp(0) with time zone NOT NULL DEFAULT now(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT now()
);