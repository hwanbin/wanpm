CREATE TYPE timesheet_status AS ENUM ('active', 'inactive', 'canceled', 'submitted', 'approved', 'rejected');

-- CREATE DOMAIN ulid AS CHAR(26)
--     CHECK (VALUE ~ '^[0-9A-HJKMNPQRSTVWXYZ]{26}$');

CREATE TABLE IF NOT EXISTS timesheet (
    id ulid NOT NULL PRIMARY KEY,
    appuser_id cuid NOT NULL,
    project_id integer NOT NULL,
    client_id integer NOT NULL,
    activity_id integer NOT NULL,
    work_date date NOT NULL,
        CONSTRAINT work_date_format CHECK (work_date = TO_DATE(work_date::TEXT, 'YYYY-MM-DD')),
    work_minutes integer NOT NULL CHECK (work_minutes > 0),
    description text NOT NULL DEFAULT '',
    status timesheet_status NOT NULL DEFAULT 'active',
    version integer NOT NULL DEFAULT 1,
    created_at timestamp(0) with time zone NOT NULL DEFAULT now(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT now(),
    FOREIGN KEY (client_id) REFERENCES client(id),
    FOREIGN KEY (project_id) REFERENCES project(project_id),
    FOREIGN KEY (appuser_id) REFERENCES appuser(id),
    FOREIGN KEY (activity_id) REFERENCES activity(id)
);

CREATE INDEX idx_timesheet_client ON timesheet (client_id);
CREATE INDEX idx_timesheet_project ON timesheet (project_id);
CREATE INDEX idx_timesheet_appuser ON timesheet (appuser_id);
CREATE INDEX idx_timesheet_activity ON timesheet (activity_id);