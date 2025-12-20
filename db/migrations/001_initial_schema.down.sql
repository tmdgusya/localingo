-- Drop tables in reverse order (to handle foreign key constraints)
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversations;

-- Optionally drop the extension (commented out as other databases might use it)
-- DROP EXTENSION IF EXISTS "pgcrypto";
