CREATE TABLE IF NOT EXISTS timesheet_project (
    timesheet_id ulid NOT NULL,
    project_id integer NOT NULL,
    PRIMARY KEY (timesheet_id, project_id),
    FOREIGN KEY (timesheet_id) REFERENCES timesheet(id) ON DELETE CASCADE,
    FOREIGN KEY (project_id) REFERENCES project(project_id) ON DELETE CASCADE
);