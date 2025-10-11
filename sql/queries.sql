-- Users
-- name: CreateUser :one
INSERT INTO users (email, password_hash, full_name)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET full_name = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- Documents
-- name: CreateDocument :one
INSERT INTO documents (user_id, filename, file_path, encrypted_key, file_size, mime_type, checksum)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetDocumentByID :one
SELECT * FROM documents WHERE id = $1;

-- name: ListDocumentsByUser :many
SELECT * FROM documents WHERE user_id = $1 ORDER BY created_at DESC;

-- name: DeleteDocument :exec
DELETE FROM documents WHERE id = $1 AND user_id = $2;

-- Shares
-- name: CreateShare :one
INSERT INTO shares (document_id, share_token, expires_at, max_access, password_hash, created_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetShareByToken :one
SELECT s.*, d.filename, d.mime_type, d.file_size, d.file_path
FROM shares s
JOIN documents d ON s.document_id = d.id
WHERE s.share_token = $1;

-- name: UpdateShareAccess :exec
UPDATE shares
SET access_count = access_count + 1
WHERE id = $1 AND (max_access = -1 OR access_count < max_access);

-- name: DeleteExpiredShares :exec
DELETE FROM shares WHERE expires_at < CURRENT_TIMESTAMP;

-- Sessions
-- name: CreateSession :one
INSERT INTO sessions (user_id, token, expires_at, ip_address, user_agent)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetSessionByToken :one
SELECT * FROM sessions WHERE token = $1;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions WHERE expires_at < CURRENT_TIMESTAMP;