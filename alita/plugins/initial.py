from alita.bot_class import Alita
from pyrogram.types import Message
from alita import LOGGER
from alita.db import (
    users_db as userdb,
    lang_db as langdb,
    rules_db as ruledb,
    blacklist_db as bldb,
    notes_db as notedb,
)


@Alita.on_message(group=-1)
async def initial_works(c: Alita, m: Message):
    try:
        if m.migrate_to_chat_id or m.migrate_from_chat_id:
            if m.migrate_to_chat_id:
                old_chat = m.chat.id
                new_chat = m.migrate_to_chat_id
            elif m.migrate_from_chat_id:
                old_chat = m.migrate_from_chat_id
                new_chat = m.chat.id

            try:
                await migrate_chat(old_chat, new_chat)
            except Exception as ef:
                LOGGER.error(ef)
                return
        else:
            userdb.update_user(
                m.from_user.id, m.from_user.username, m.chat.id, m.chat.title
            )
            if m.reply_to_message:
                userdb.update_user(
                    m.reply_to_message.from_user.id,
                    m.reply_to_message.from_user.username,
                    m.chat.id,
                    m.chat.title,
                )
            if m.forward_from:
                userdb.update_user(m.forward_from.id, m.forward_from.username)
    except AttributeError:
        pass  # Skip attribute errors!
    return


async def migrate_chat(old_chat, new_chat):
    LOGGER.info(f"Migrating from {str(old_chat)} to {str(new_chat)}")
    userdb.migrate_chat(old_chat, new_chat)
    langdb.migrate_chat(old_chat, new_chat)
    ruledb.migrate_chat(old_chat, new_chat)
    bldb.migrate_chat(old_chat, new_chat)
    notedb.migrate_chat(old_chat, new_chat)
    LOGGER.info("Successfully migrated!")
