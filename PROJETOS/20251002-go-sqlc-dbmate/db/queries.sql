-- name: CreateUser :one
INSERT INTO users (
    username
) VALUES (
    $1
)
RETURNING *;

-- name: GetUserAudits :many
SELECT * from users_audit_logs;
