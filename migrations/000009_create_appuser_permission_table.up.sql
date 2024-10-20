CREATE TABLE IF NOT EXISTS appuser_permission (
    user_internal_id integer NOT NULL,
    permission_internal_id integer NOT NULL,
    PRIMARY KEY (user_internal_id, permission_internal_id),
    FOREIGN KEY (user_internal_id) REFERENCES appuser(internal_id) ON DELETE CASCADE,
    FOREIGN KEY (permission_internal_id) REFERENCES permission(internal_id) ON DELETE CASCADE
);