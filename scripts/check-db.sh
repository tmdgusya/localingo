#!/bin/bash

echo "=== Checking Chat History Database ==="
echo ""

echo "📊 Conversations:"
docker exec -it localingo-postgres psql -U localingo_user -d localingo -c "
SELECT
    id,
    title,
    created_at,
    (SELECT COUNT(*) FROM messages WHERE conversation_id = conversations.id) as message_count
FROM conversations
ORDER BY created_at DESC;
"

echo ""
echo "💬 Messages:"
docker exec -it localingo-postgres psql -U localingo_user -d localingo -c "
SELECT
    m.id,
    c.id as conversation_id,
    m.role,
    LEFT(m.content, 50) as content_preview,
    m.model,
    m.created_at,
    m.completion_time_ms
FROM messages m
JOIN conversations c ON m.conversation_id = c.id
ORDER BY m.created_at DESC
LIMIT 10;
"

echo ""
echo "📈 Summary:"
docker exec -it localingo-postgres psql -U localingo_user -d localingo -c "
SELECT
    (SELECT COUNT(*) FROM conversations) as total_conversations,
    (SELECT COUNT(*) FROM messages) as total_messages,
    (SELECT COUNT(*) FROM messages WHERE role = 'user') as user_messages,
    (SELECT COUNT(*) FROM messages WHERE role = 'assistant') as assistant_messages;
"
