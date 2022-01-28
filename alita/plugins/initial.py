from pyrogram import filters
from pyrogram.errors import RPCError
from pyrogram.types import Message

from alita import LOGGER
from alita.bot_class import Alita
from alita.database.approve_db import Approve
from alita.database.blacklist_db import Blacklist
from alita.database.chats_db import Chats
from alita.database.disable_db import Disabling
from alita.database.filters_db import Filters
from alita.database.greetings_db import Greetings
from alita.database.lang_db import Langs
from alita.database.notes_db import Notes, NotesSettings
from alita.database.pins_db import Pins
from alita.database.reporting_db import Reporting
from alita.database.rules_db import Rules
from alita.database.users_db import Users


@Alita.on_message(filters.group, group=4)
async def initial_works(_, m: Message):
    chatdb = Chats(m.chat.id)
    try:
        if m.migrate_to_chat_id or m.migrate_from_chat_id:
            new_chat = m.migrate_to_chat_id or m.chat.id
            try:
                await migrate_chat(m, new_chat)
            except RPCError as ef:
                LOGGER.error(ef)
                return
        elif m.reply_to_message and not m.forward_from:
            chatdb.update_chat(
                m.chat.title,
                m.reply_to_message.from_user.id,
            )
            Users(m.reply_to_message.from_user.id).update_user(
                (
                    f"{m.reply_to_message.from_user.first_name} {m.reply_to_message.from_user.last_name}"
                    if m.reply_to_message.from_user.last_name
                    else m.reply_to_message.from_user.first_name
                ),
                m.reply_to_message.from_user.username,
            )
        elif m.forward_from and not m.reply_to_message:
            chatdb.update_chat(
                m.chat.title,
                m.forward_from.id,
            )
            Users(m.forward_from.id).update_user(
                (
                    f"{m.forward_from.first_name} {m.forward_from.last_name}"
                    if m.forward_from.last_name
                    else m.forward_from.first_name
                ),
                m.forward_from.username,
            )
        elif m.reply_to_message:
            chatdb.update_chat(
                m.chat.title,
                m.reply_to_message.forward_from.id,
            )
            Users(m.forward_from.id).update_user(
                (
                    f"{m.reply_to_message.forward_from.first_name} {m.reply_to_message.forward_from.last_name}"
                    if m.reply_to_message.forward_from.last_name
                    else m.reply_to_message.forward_from.first_name
                ),
                m.forward_from.username,
            )
        else:
            chatdb.update_chat(m.chat.title, m.from_user.id)
            Users(m.from_user.id).update_user(
                (
                    f"{m.from_user.first_name} {m.from_user.last_name}"
                    if m.from_user.last_name
                    else m.from_user.first_name
                ),
                m.from_user.username,
            )
    except AttributeError:
        pass  # Skip attribute errors!
    return


async def migrate_chat(m: Message, new_chat: int) -> None:
    LOGGER.info(f"Migrating from {m.chat.id} to {new_chat}...")
    langdb = Langs(m.chat.id)
    notedb = Notes()
    gdb = Greetings(m.chat.id)
    ruledb = Rules(m.chat.id)
    userdb = Users(m.chat.id)
    chatdb = Chats(m.chat.id)
    bldb = Blacklist(m.chat.id)
    approvedb = Approve(m.chat.id)
    reportdb = Reporting(m.chat.id)
    notes_settings = NotesSettings()
    pins_db = Pins(m.chat.id)
    fldb = Filters()
    disabl = Disabling(m.chat.id)
    disabl.migrate_chat(new_chat)
    gdb.migrate_chat(new_chat)
    chatdb.migrate_chat(new_chat)
    userdb.migrate_chat(new_chat)
    langdb.migrate_chat(new_chat)
    ruledb.migrate_chat(new_chat)
    bldb.migrate_chat(new_chat)
    notedb.migrate_chat(m.chat.id, new_chat)
    approvedb.migrate_chat(new_chat)
    reportdb.migrate_chat(new_chat)
    notes_settings.migrate_chat(m.chat.id, new_chat)
    pins_db.migrate_chat(new_chat)
    fldb.migrate_chat(m.chat.id, new_chat)
    LOGGER.info(f"Successfully migrated from {m.chat.id} to {new_chat}!")
