-- Add bot_id column to all tables to support bot cloning
-- Bot ID 1 = main Alita bot, other IDs = cloned bot instances
-- The bot_id is extracted from Telegram bot token format: {BOT_ID}:{SECRET}

-- 1. Add bot_id column to all existing tables with DEFAULT 1 (main bot)
-- This ensures backward compatibility and all existing data belongs to main bot

-- Core tables
ALTER TABLE admin ADD COLUMN IF NOT EXISTS bot_id BIGINT DEFAULT 1 NOT NULL;
ALTER TABLE antiflood_settings ADD COLUMN IF NOT EXISTS bot_id BIGINT DEFAULT 1 NOT NULL;
ALTER TABLE blacklists ADD COLUMN IF NOT EXISTS bot_id BIGINT DEFAULT 1 NOT NULL;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS bot_id BIGINT DEFAULT 1 NOT NULL;
ALTER TABLE chats ADD COLUMN IF NOT EXISTS bot_id BIGINT DEFAULT 1 NOT NULL;
ALTER TABLE chat_users ADD COLUMN IF NOT EXISTS bot_id BIGINT DEFAULT 1 NOT NULL;
ALTER TABLE connection ADD COLUMN IF NOT EXISTS bot_id BIGINT DEFAULT 1 NOT NULL;
ALTER TABLE connection_settings ADD COLUMN IF NOT EXISTS bot_id BIGINT DEFAULT 1 NOT NULL;
ALTER TABLE devs ADD COLUMN IF NOT EXISTS bot_id BIGINT DEFAULT 1 NOT NULL;
ALTER TABLE disable ADD COLUMN IF NOT EXISTS bot_id BIGINT DEFAULT 1 NOT NULL;
ALTER TABLE filters ADD COLUMN IF NOT EXISTS bot_id BIGINT DEFAULT 1 NOT NULL;
ALTER TABLE greetings ADD COLUMN IF NOT EXISTS bot_id BIGINT DEFAULT 1 NOT NULL;
ALTER TABLE locks ADD COLUMN IF NOT EXISTS bot_id BIGINT DEFAULT 1 NOT NULL;
ALTER TABLE notes ADD COLUMN IF NOT EXISTS bot_id BIGINT DEFAULT 1 NOT NULL;
ALTER TABLE notes_settings ADD COLUMN IF NOT EXISTS bot_id BIGINT DEFAULT 1 NOT NULL;
ALTER TABLE pins ADD COLUMN IF NOT EXISTS bot_id BIGINT DEFAULT 1 NOT NULL;
ALTER TABLE report_chat_settings ADD COLUMN IF NOT EXISTS bot_id BIGINT DEFAULT 1 NOT NULL;
ALTER TABLE report_user_settings ADD COLUMN IF NOT EXISTS bot_id BIGINT DEFAULT 1 NOT NULL;
ALTER TABLE rules ADD COLUMN IF NOT EXISTS bot_id BIGINT DEFAULT 1 NOT NULL;
ALTER TABLE users ADD COLUMN IF NOT EXISTS bot_id BIGINT DEFAULT 1 NOT NULL;
ALTER TABLE warns_settings ADD COLUMN IF NOT EXISTS bot_id BIGINT DEFAULT 1 NOT NULL;
ALTER TABLE warns_users ADD COLUMN IF NOT EXISTS bot_id BIGINT DEFAULT 1 NOT NULL;

-- Captcha tables
ALTER TABLE captcha_settings ADD COLUMN IF NOT EXISTS bot_id BIGINT DEFAULT 1 NOT NULL;
ALTER TABLE captcha_attempts ADD COLUMN IF NOT EXISTS bot_id BIGINT DEFAULT 1 NOT NULL;

-- 2. Create new table to store cloned bot instances and their tokens
CREATE TABLE IF NOT EXISTS bot_instances (
    id SERIAL PRIMARY KEY,
    bot_id BIGINT NOT NULL UNIQUE,
    owner_id BIGINT NOT NULL,
    token_hash VARCHAR(64) NOT NULL,  -- Hashed token for security
    bot_username VARCHAR(255),
    bot_name VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_activity TIMESTAMP,
    webhook_url VARCHAR(500),
    
    -- Foreign key to main users table (owner must exist)
    CONSTRAINT fk_bot_instances_owner FOREIGN KEY (owner_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Create indexes for bot_instances
CREATE UNIQUE INDEX IF NOT EXISTS idx_bot_instances_bot_id ON bot_instances(bot_id);
CREATE INDEX IF NOT EXISTS idx_bot_instances_owner_id ON bot_instances(owner_id);
CREATE INDEX IF NOT EXISTS idx_bot_instances_active ON bot_instances(is_active, bot_id);

-- Add trigger to update updated_at timestamp for bot_instances
CREATE OR REPLACE FUNCTION update_bot_instances_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_bot_instances_updated_at
    BEFORE UPDATE ON bot_instances
    FOR EACH ROW
    EXECUTE FUNCTION update_bot_instances_updated_at();

-- 3. Update existing unique constraints to include bot_id where appropriate
-- This ensures data isolation between bot instances

-- Drop old unique constraints that need to include bot_id
ALTER TABLE admin DROP CONSTRAINT IF EXISTS idx_admin_chat_id;
ALTER TABLE antiflood_settings DROP CONSTRAINT IF EXISTS idx_antiflood_settings_chat_id;
ALTER TABLE channels DROP CONSTRAINT IF EXISTS idx_channels_chat_id;  
ALTER TABLE chats DROP CONSTRAINT IF EXISTS idx_chats_chat_id;
ALTER TABLE connection_settings DROP CONSTRAINT IF EXISTS idx_connection_settings_chat_id;
ALTER TABLE devs DROP CONSTRAINT IF EXISTS idx_devs_user_id;
ALTER TABLE greetings DROP CONSTRAINT IF EXISTS idx_greetings_chat_id;
ALTER TABLE notes_settings DROP CONSTRAINT IF EXISTS idx_notes_settings_chat_id;
ALTER TABLE pins DROP CONSTRAINT IF EXISTS idx_pins_chat_id;
ALTER TABLE report_chat_settings DROP CONSTRAINT IF EXISTS idx_report_chat_settings_chat_id;
ALTER TABLE report_user_settings DROP CONSTRAINT IF EXISTS idx_report_user_settings_user_id;
ALTER TABLE rules DROP CONSTRAINT IF EXISTS idx_rules_chat_id;
ALTER TABLE users DROP CONSTRAINT IF EXISTS idx_users_user_id;
ALTER TABLE warns_settings DROP CONSTRAINT IF EXISTS idx_warns_settings_chat_id;
ALTER TABLE captcha_settings DROP CONSTRAINT IF EXISTS uk_captcha_settings_chat_id;

-- Create new unique constraints that include bot_id for proper isolation
CREATE UNIQUE INDEX IF NOT EXISTS idx_admin_bot_chat ON admin(bot_id, chat_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_antiflood_settings_bot_chat ON antiflood_settings(bot_id, chat_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_channels_bot_chat ON channels(bot_id, chat_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_chats_bot_chat ON chats(bot_id, chat_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_connection_settings_bot_chat ON connection_settings(bot_id, chat_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_devs_bot_user ON devs(bot_id, user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_greetings_bot_chat ON greetings(bot_id, chat_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_notes_settings_bot_chat ON notes_settings(bot_id, chat_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_pins_bot_chat ON pins(bot_id, chat_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_report_chat_settings_bot_chat ON report_chat_settings(bot_id, chat_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_report_user_settings_bot_user ON report_user_settings(bot_id, user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_rules_bot_chat ON rules(bot_id, chat_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_bot_user ON users(bot_id, user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_warns_settings_bot_chat ON warns_settings(bot_id, chat_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_captcha_settings_bot_chat ON captcha_settings(bot_id, chat_id);

-- Update composite primary key for chat_users to include bot_id
ALTER TABLE chat_users DROP CONSTRAINT IF EXISTS chat_users_pkey;
ALTER TABLE chat_users ADD CONSTRAINT chat_users_pkey PRIMARY KEY (bot_id, chat_id, user_id);

-- 4. Create performance indexes for bot_id queries
CREATE INDEX IF NOT EXISTS idx_admin_bot_id ON admin(bot_id);
CREATE INDEX IF NOT EXISTS idx_antiflood_settings_bot_id ON antiflood_settings(bot_id);
CREATE INDEX IF NOT EXISTS idx_blacklists_bot_id ON blacklists(bot_id);
CREATE INDEX IF NOT EXISTS idx_channels_bot_id ON channels(bot_id);
CREATE INDEX IF NOT EXISTS idx_chats_bot_id ON chats(bot_id);
CREATE INDEX IF NOT EXISTS idx_chat_users_bot_id ON chat_users(bot_id);
CREATE INDEX IF NOT EXISTS idx_connection_bot_id ON connection(bot_id);
CREATE INDEX IF NOT EXISTS idx_connection_settings_bot_id ON connection_settings(bot_id);
CREATE INDEX IF NOT EXISTS idx_devs_bot_id ON devs(bot_id);
CREATE INDEX IF NOT EXISTS idx_disable_bot_id ON disable(bot_id);
CREATE INDEX IF NOT EXISTS idx_filters_bot_id ON filters(bot_id);
CREATE INDEX IF NOT EXISTS idx_greetings_bot_id ON greetings(bot_id);
CREATE INDEX IF NOT EXISTS idx_locks_bot_id ON locks(bot_id);
CREATE INDEX IF NOT EXISTS idx_notes_bot_id ON notes(bot_id);
CREATE INDEX IF NOT EXISTS idx_notes_settings_bot_id ON notes_settings(bot_id);
CREATE INDEX IF NOT EXISTS idx_pins_bot_id ON pins(bot_id);
CREATE INDEX IF NOT EXISTS idx_report_chat_settings_bot_id ON report_chat_settings(bot_id);
CREATE INDEX IF NOT EXISTS idx_report_user_settings_bot_id ON report_user_settings(bot_id);
CREATE INDEX IF NOT EXISTS idx_rules_bot_id ON rules(bot_id);
CREATE INDEX IF NOT EXISTS idx_users_bot_id ON users(bot_id);
CREATE INDEX IF NOT EXISTS idx_warns_settings_bot_id ON warns_settings(bot_id);
CREATE INDEX IF NOT EXISTS idx_warns_users_bot_id ON warns_users(bot_id);
CREATE INDEX IF NOT EXISTS idx_captcha_settings_bot_id ON captcha_settings(bot_id);
CREATE INDEX IF NOT EXISTS idx_captcha_attempts_bot_id ON captcha_attempts(bot_id);

-- 5. Add comments explaining the bot_id column
COMMENT ON COLUMN admin.bot_id IS 'Bot instance ID extracted from Telegram token - 1 for main Alita bot, others for clones';
COMMENT ON COLUMN antiflood_settings.bot_id IS 'Bot instance ID extracted from Telegram token - 1 for main Alita bot, others for clones';
COMMENT ON COLUMN blacklists.bot_id IS 'Bot instance ID extracted from Telegram token - 1 for main Alita bot, others for clones';
COMMENT ON COLUMN channels.bot_id IS 'Bot instance ID extracted from Telegram token - 1 for main Alita bot, others for clones';
COMMENT ON COLUMN chats.bot_id IS 'Bot instance ID extracted from Telegram token - 1 for main Alita bot, others for clones';
COMMENT ON COLUMN chat_users.bot_id IS 'Bot instance ID extracted from Telegram token - 1 for main Alita bot, others for clones';
COMMENT ON COLUMN connection.bot_id IS 'Bot instance ID extracted from Telegram token - 1 for main Alita bot, others for clones';
COMMENT ON COLUMN connection_settings.bot_id IS 'Bot instance ID extracted from Telegram token - 1 for main Alita bot, others for clones';
COMMENT ON COLUMN devs.bot_id IS 'Bot instance ID extracted from Telegram token - 1 for main Alita bot, others for clones';
COMMENT ON COLUMN disable.bot_id IS 'Bot instance ID extracted from Telegram token - 1 for main Alita bot, others for clones';
COMMENT ON COLUMN filters.bot_id IS 'Bot instance ID extracted from Telegram token - 1 for main Alita bot, others for clones';
COMMENT ON COLUMN greetings.bot_id IS 'Bot instance ID extracted from Telegram token - 1 for main Alita bot, others for clones';
COMMENT ON COLUMN locks.bot_id IS 'Bot instance ID extracted from Telegram token - 1 for main Alita bot, others for clones';
COMMENT ON COLUMN notes.bot_id IS 'Bot instance ID extracted from Telegram token - 1 for main Alita bot, others for clones';
COMMENT ON COLUMN notes_settings.bot_id IS 'Bot instance ID extracted from Telegram token - 1 for main Alita bot, others for clones';
COMMENT ON COLUMN pins.bot_id IS 'Bot instance ID extracted from Telegram token - 1 for main Alita bot, others for clones';
COMMENT ON COLUMN report_chat_settings.bot_id IS 'Bot instance ID extracted from Telegram token - 1 for main Alita bot, others for clones';
COMMENT ON COLUMN report_user_settings.bot_id IS 'Bot instance ID extracted from Telegram token - 1 for main Alita bot, others for clones';
COMMENT ON COLUMN rules.bot_id IS 'Bot instance ID extracted from Telegram token - 1 for main Alita bot, others for clones';
COMMENT ON COLUMN users.bot_id IS 'Bot instance ID extracted from Telegram token - 1 for main Alita bot, others for clones';
COMMENT ON COLUMN warns_settings.bot_id IS 'Bot instance ID extracted from Telegram token - 1 for main Alita bot, others for clones';
COMMENT ON COLUMN warns_users.bot_id IS 'Bot instance ID extracted from Telegram token - 1 for main Alita bot, others for clones';
COMMENT ON COLUMN captcha_settings.bot_id IS 'Bot instance ID extracted from Telegram token - 1 for main Alita bot, others for clones';
COMMENT ON COLUMN captcha_attempts.bot_id IS 'Bot instance ID extracted from Telegram token - 1 for main Alita bot, others for clones';

-- Add comment on the bot_instances table
COMMENT ON TABLE bot_instances IS 'Stores cloned bot instances with encrypted tokens and ownership info';