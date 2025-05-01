CREATE TABLE IF NOT EXISTS timesheet_appuser (
    timesheet_id ulid NOT NULL,
    appuser_id cuid NOT NULL,
    PRIMARY KEY (timesheet_id, appuser_id),
    FOREIGN KEY (timesheet_id) REFERENCES timesheet(id) ON DELETE CASCADE,
    FOREIGN KEY (appuser_id) REFERENCES appuser(id) ON DELETE CASCADE
);