CREATE TABLE IF NOT EXISTS project (
    internal_id serial PRIMARY KEY,
    project_id integer UNIQUE NOT NULL,
    proposal_id text UNIQUE NOT NULL,
    name text,
    status text,
    coordinates geometry,
    images text[],
    version integer NOT NULL DEFAULT 1,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_geom ON project USING GIST(coordinates);