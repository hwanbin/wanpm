CREATE TABLE IF NOT EXISTS project (
    internal_id serial PRIMARY KEY,
    project_id integer UNIQUE NOT NULL,
    proposal_id text UNIQUE NOT NULL,
    name text,
    status text,
    feature jsonb,
    images text[],
    note text,
    version integer NOT NULL DEFAULT 1,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_project_name ON project (name);
CREATE INDEX idx_project_status ON project (status);
CREATE INDEX idx_full_address ON project ((feature->'properties'->>'full_address'));
CREATE INDEX idx_geom ON project USING GIST(
    ST_SetSRID(
        ST_MakePoint(
            (feature->'geometry'->'coordinates'->>0)::float,
            (feature->'geometry'->'coordinates'->>1)::float
        ),
        4326
    )
);