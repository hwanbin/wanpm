CREATE TABLE IF NOT EXISTS assignment (
    employee_id cuid NOT NULL,
    role_id integer NOT NULL,
    project_id integer NOT NULL,
    PRIMARY KEY (employee_id, role_id, project_id),
    FOREIGN KEY (employee_id) REFERENCES appuser(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES role(id) ON DELETE CASCADE,
    FOREIGN KEY (project_id) REFERENCES project(project_id) ON DELETE CASCADE
);