-- Drop the log_messages table and its indexes
-- Logs are now stored in-memory like tasks

DROP INDEX IF EXISTS idx_log_messages_timestamp;
DROP INDEX IF EXISTS idx_log_messages_level;
DROP TABLE IF EXISTS log_messages;
