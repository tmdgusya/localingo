-- name: CreateMessage :one
INSERT INTO messages (conversation_id, role, content, model, metadata, token_count, completion_time_ms)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetMessage :one
SELECT * FROM messages
WHERE id = $1;

-- name: ListMessagesByConversation :many
SELECT * FROM messages
WHERE conversation_id = $1
ORDER BY created_at ASC;

-- name: ListMessagesByConversationPaginated :many
SELECT * FROM messages
WHERE conversation_id = $1
ORDER BY created_at ASC
LIMIT $2 OFFSET $3;

-- name: DeleteMessage :exec
DELETE FROM messages
WHERE id = $1;

-- name: DeleteMessagesByConversation :exec
DELETE FROM messages
WHERE conversation_id = $1;

-- name: GetConversationMessageCount :one
SELECT COUNT(*) FROM messages
WHERE conversation_id = $1;

-- name: GetLatestMessageByConversation :one
SELECT * FROM messages
WHERE conversation_id = $1
ORDER BY created_at DESC
LIMIT 1;
