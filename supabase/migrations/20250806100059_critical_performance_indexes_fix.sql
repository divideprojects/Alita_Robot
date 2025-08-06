-- =====================================================
-- CRITICAL PERFORMANCE INDEXES MIGRATION
-- =====================================================
-- This migration addresses severe performance issues identified in query performance reports
-- Priority 1: Most critical indexes for high-frequency queries

-- =====================================================
-- 1. LOCKS TABLE - Most Critical (319,183 calls)
-- =====================================================
-- Composite index for chat_id and lock_type lookups
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_locks_chat_lock_lookup 
ON locks(chat_id, lock_type) 
WHERE locked = true;

-- Covering index to avoid table lookups
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_locks_covering 
ON locks(chat_id, lock_type) 
INCLUDE (locked, id);

-- =====================================================
-- 2. USERS TABLE - High Impact (61,067 calls)
-- =====================================================
-- Primary lookup index
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_user_id_active 
ON users(user_id) 
WHERE user_id IS NOT NULL;

-- Covering index for common fields
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_covering 
ON users(user_id) 
INCLUDE (username, name, language);

-- Partial index for recently active users
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_active_partial 
ON users(user_id) 
WHERE updated_at > CURRENT_DATE - INTERVAL '30 days';

-- =====================================================
-- 3. CHATS TABLE - High Impact (123,242 calls)  
-- =====================================================
-- Primary lookup index for active chats
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_chats_chat_id_active 
ON chats(chat_id) 
WHERE is_inactive = false;

-- Covering index for common fields
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_chats_covering 
ON chats(chat_id) 
INCLUDE (chat_name, language, users, is_inactive);

-- GIN index for users JSONB field if it exists
DO $$ 
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'chats' 
        AND column_name = 'users' 
        AND data_type = 'jsonb'
    ) THEN
        CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_chats_users_gin 
        ON chats USING GIN (users) 
        WHERE users IS NOT NULL;
    END IF;
END $$;

-- =====================================================
-- 4. ANTIFLOOD SETTINGS - Medium Impact (58,166 calls)
-- =====================================================
-- Index for active antiflood settings
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_antiflood_chat_active 
ON antiflood_settings(chat_id) 
WHERE "limit" > 0;

-- =====================================================
-- 5. FILTERS - Medium Impact (33,898 calls)
-- =====================================================
-- Optimized index with included columns
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_filters_chat_optimized 
ON filters(chat_id) 
INCLUDE (keyword, filter_reply);

-- =====================================================
-- 6. BLACKLISTS - Easy Win (33,474 calls)
-- =====================================================
-- Optimized index with included columns
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_blacklists_chat_word_optimized 
ON blacklists(chat_id) 
INCLUDE (word, action);

-- =====================================================
-- 7. CHANNELS - For Frequent Updates (17,907 calls)
-- =====================================================
-- Index for update operations
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_channels_chat_update 
ON channels(chat_id) 
INCLUDE (channel_id, updated_at);

-- =====================================================
-- 8. GREETINGS - For Message Handling (1,407 calls)
-- =====================================================
-- Partial index for enabled greetings
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_greetings_chat_enabled 
ON greetings(chat_id) 
WHERE welcome_enabled = true OR goodbye_enabled = true;

-- =====================================================
-- 9. ADDITIONAL PERFORMANCE INDEXES
-- =====================================================

-- Warns users composite index
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_warns_users_composite 
ON warns_users(user_id, chat_id) 
INCLUDE (num_warns, warns)
WHERE num_warns > 0;

-- Notes lookup index
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_notes_chat_name 
ON notes(chat_id, note_name);

-- Admin settings lookup
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_admin_settings_chat 
ON admin_settings(chat_id);

-- Pins lookup index
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_pins_chat 
ON pins(chat_id);

-- Cleanup indexes for old data
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_cleanup 
ON users(updated_at DESC) 
WHERE updated_at < CURRENT_DATE - INTERVAL '90 days';

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_chats_cleanup 
ON chats(updated_at DESC) 
WHERE is_inactive = true;

-- =====================================================
-- 10. UPDATE TABLE STATISTICS
-- =====================================================
-- Update statistics for better query planning
ANALYZE users;
ANALYZE chats;
ANALYZE locks;
ANALYZE antiflood_settings;
ANALYZE filters;
ANALYZE blacklists;
ANALYZE channels;
ANALYZE greetings;
ANALYZE warns_users;
ANALYZE notes;
ANALYZE admin_settings;
ANALYZE pins;

-- Set higher statistics targets for frequently queried columns
ALTER TABLE users ALTER COLUMN user_id SET STATISTICS 1000;
ALTER TABLE chats ALTER COLUMN chat_id SET STATISTICS 1000;
ALTER TABLE locks ALTER COLUMN chat_id SET STATISTICS 500;
ALTER TABLE antiflood_settings ALTER COLUMN chat_id SET STATISTICS 500;
ALTER TABLE filters ALTER COLUMN chat_id SET STATISTICS 500;

-- =====================================================
-- VALIDATION QUERIES
-- =====================================================
-- These queries help verify that indexes are created and being used

-- List all new indexes created
SELECT schemaname, tablename, indexname, indexdef
FROM pg_indexes
WHERE indexname LIKE 'idx_%_chat_%' 
   OR indexname LIKE 'idx_%_covering'
   OR indexname LIKE 'idx_%_active%'
   OR indexname LIKE 'idx_%_cleanup'
ORDER BY tablename, indexname;

-- Check index usage statistics (run after some time)
-- SELECT 
--     schemaname,
--     tablename,
--     indexname,
--     idx_scan,
--     idx_tup_read,
--     idx_tup_fetch
-- FROM pg_stat_user_indexes
-- WHERE schemaname = 'public'
-- ORDER BY idx_scan DESC;
