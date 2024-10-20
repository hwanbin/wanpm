CREATE TABLE IF NOT EXISTS client (
    internal_id serial PRIMARY KEY,
    name text NOT NULL,
    address text,
    logo_url text,
    note text,
    version integer NOT NULL DEFAULT 1,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);