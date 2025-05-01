CREATE TABLE IF NOT EXISTS appuser_permission (
    user_id cuid NOT NULL,
    permission_internal_id integer NOT NULL,
    PRIMARY KEY (user_id, permission_internal_id),
    FOREIGN KEY (user_id) REFERENCES appuser(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_internal_id) REFERENCES permission(internal_id) ON DELETE CASCADE
);