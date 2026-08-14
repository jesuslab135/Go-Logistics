-- BR-05: denylist for revoked refresh-token jtis, kept until token expiry.
-- Rows older than their expires_at are dead weight and are pruned
-- opportunistically on logout.
CREATE TABLE revoked_token (
    jti         varchar(64) PRIMARY KEY,
    employee_id bigint      NOT NULL,
    expires_at  timestamptz NOT NULL
);

CREATE INDEX idx_revoked_token_expires_at ON revoked_token (expires_at);
