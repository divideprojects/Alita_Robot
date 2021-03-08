from pyrogram import filters
from pyrogram.types import Message

from alita import PREFIX_HANDLER
from alita.bot_class import Alita
from alita.database.spam_protect_db import SpamProtect
from alita.utils.custom_filters import admin_filter

# Initialise
db = SpamProtect()

__PLUGIN__ = "Spam Protect"
__help__ = """
Spammers joining your group?
No problem at all, here you can find all the options to protect your groups from raids and scums!

**Admin Only:**
 × /cas <on/off>: Turns on or off the CAS check for group.
 × /underattack <on/off>: Kick all the new users on entry!
"""


@Alita.on_message(
    filters.command("cas", PREFIX_HANDLER) & filters.group & admin_filter,
)
async def cas_protect(_, m: Message):
    get_cas = db.get_cas_status(m.chat.id)

    if len(m.text.split()) == 2:
        new_s = m.text.split(None, 1)[1]
        if new_s.lower() in ("yes", "on", "true"):
            yn = True
        elif new_s.lower() in ("no", "off", "false"):
            yn = False
        else:
            await m.reply_text(
                ("Please use an option out of:\non, yes, true or no, off, false"),
            )
            return
        db.set_cas_status(m.chat.id, yn)
        await m.reply_text(f"Set CAS Status to <code>{new_s}</code>")
    else:
        await m.reply_text(f"Your current CAS Setting is: <b>{get_cas}</b>")

    return


@Alita.on_message(
    filters.command("underattack", PREFIX_HANDLER) & filters.group & admin_filter,
)
async def underattack(_, m: Message):
    get_a = db.get_attack_status(m.chat.id)

    if len(m.text.split()) == 2:
        new_s = m.text.split(None, 1)[1]
        if new_s.lower() in ("yes", "on", "true"):
            yn = True
        elif new_s.lower() in ("no", "off", "false"):
            yn = False
        else:
            await m.reply_text(
                ("Please use an option out of:\non, yes, true or no, off, false"),
            )
            return
        db.set_attack_status(m.chat.id, yn)
        await m.reply_text(f"Set UnderAttack Status to <code>{new_s}</code>")
    else:
        await m.reply_text(f"Your current underAttack Status is: <b>{get_a}</b>")

    return
