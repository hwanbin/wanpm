DROP INDEX IF EXISTS idx_timesheet_activity;
DROP INDEX IF EXISTS idx_timesheet_appuser;
DROP INDEX IF EXISTS idx_timesheet_project;
DROP INDEX IF EXISTS idx_timesheet_client;

DROP TABLE IF EXISTS timesheet;

-- DROP DOMAIN IF EXISTS ulid;
DROP TYPE IF EXISTS timesheet_status;