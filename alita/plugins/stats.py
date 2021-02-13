from pyrogram import filters
from pyrogram.types import Message
from alita import Alita, DEV_PREFIX_HANDLER
from alita.utils.custom_filters import dev_filter
from alita.db import (
    users_db as userdb,
    blacklist_db as bldb,
    rules_db as rulesdb,
    notes_db as notesdb,
    antispam_db as gbandb,
)


@Alita.on_message(filters.command("stats", DEV_PREFIX_HANDLER) & dev_filter)
async def get_stats(c: Alita, m: Message):
    sm = await m.reply_text("**__Fetching Stats...__**")
    rply = (
        f"<b>Users:</b> <code>{userdb.num_users()}</code> in <code>{userdb.num_chats()}</code> chats\n"
        f"<b>Blacklists:</b> <code>{bldb.num_blacklist_filters()}</code> in <code>{bldb.num_blacklist_filter_chats()}</code> chats\n"
        f"<b>Rules:</b> Set in <code>{rulesdb.num_chats()}</code> chats\n"
        f"<b>Notes:</b> <code>{notesdb.num_notes_all()}</code> in <code>{notesdb.all_notes_chats()}</code>\n"
        f"<b>Globally Banned Users:</b> <code>{gbandb.num_gbanned_users()}</code>\n"
    )
    await sm.edit_text(rply, parse_mode="html")

    return
