-- name: CreateConversation :one
INSERT INTO conversations (title, metadata)
VALUES ($1, $2)
RETURNING *;

-- name: GetConversation :one
SELECT * FROM conversations
WHERE id = $1;

-- name: ListConversations :many
SELECT * FROM conversations
ORDER BY updated_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateConversation :one
UPDATE conversations
SET title = $2, updated_at = NOW(), metadata = $3
WHERE id = $1
RETURNING *;

-- name: DeleteConversation :exec
DELETE FROM conversations
WHERE id = $1;

-- name: UpdateConversationTimestamp :exec
UPDATE conversations
SET updated_at = NOW()
WHERE id = $1;
