CREATE TABLE IF NOT EXISTS proposal (
    internal_id serial PRIMARY KEY,
    project_id text UNIQUE NOT NULL,
    version integer NOT NULL DEFAULT 1,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);