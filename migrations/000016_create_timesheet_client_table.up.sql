CREATE TABLE IF NOT EXISTS timesheet_client (
    timesheet_id ulid NOT NULL,
    client_id integer NOT NULL,
    PRIMARY KEY (timesheet_id, client_id),
    FOREIGN KEY (timesheet_id) REFERENCES timesheet(id) ON DELETE CASCADE,
    FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE
);