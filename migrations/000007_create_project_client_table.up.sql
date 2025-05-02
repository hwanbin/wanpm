CREATE TABLE IF NOT EXISTS project_client (
    project_internal_id integer NOT NULL,
    client_id integer NOT NULL,
    PRIMARY KEY (project_internal_id, client_id),
    FOREIGN KEY (project_internal_id) REFERENCES project(internal_id) ON DELETE CASCADE,
    FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE
);