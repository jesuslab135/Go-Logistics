-- name: RevokeRefreshToken :exec
INSERT INTO revoked_token (jti, employee_id, expires_at)
VALUES ($1, $2, $3)
ON CONFLICT (jti) DO NOTHING;

-- name: IsRefreshTokenRevoked :one
SELECT EXISTS(SELECT 1 FROM revoked_token WHERE jti = $1)::boolean AS revoked;

-- name: DeleteExpiredRevokedTokens :exec
DELETE FROM revoked_token WHERE expires_at < now();
