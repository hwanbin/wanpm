CREATE TABLE IF NOT EXISTS project_appuser (
    project_internal_id integer NOT NULL,
    appuser_id cuid NOT NULL,
    PRIMARY KEY (project_internal_id, appuser_id),
    FOREIGN KEY (project_internal_id) REFERENCES project(internal_id) ON DELETE CASCADE,
    FOREIGN KEY (appuser_id) REFERENCES appuser(id) ON DELETE CASCADE
);