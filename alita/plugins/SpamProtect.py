from pyrogram import filters
from pyrogram.types import Message

from alita import (
    LOGGER,
    PREFIX_HANDLER,
    SUPPORT_GROUP,
    SUPPORT_STAFF,
)
from alita.bot_class import Alita
from alita.tr_engine import tlang
from alita.utils.custom_filters import admin_filter


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
async def cas_protect(c: Alita, m: Message):
    get_cas = await db.get_cas_status(m.chat.id)

    if len(m.text.split()) == 2:
        new_s = m.text.split(None, 1)[1]
        if new_s.lower() in ("yes", 'on', 'true'):
            yn = True
        else:
            yn = False
        await db.set_cas_status(m.chat.id, yn)
        await m.reply_text(f"Set CAS Status to {new_s}")
    else:
        await m.reply_text(f"Your current CAS Setting is: {get_cas}")

    return


@Alita.on_message(
    filters.command("underattack", PREFIX_HANDLER) & filters.group & admin_filter,
)
async def underattack(c: Alita, m: Message):
    get_a = await db.get_attack_status(m.chat.id)

    if len(m.text.split()) == 2:
        new_s = m.text.split(None, 1)[1]
        if new_s.lower() in ("yes", 'on', 'true'):
            yn = True
        else:
            yn = False
        await db.set_attack_status(m.chat.id, yn)
        await m.reply_text(f"Set underAttack Status to {new_s}")
    else:
        await m.reply_text(f"Your current underAttack Status is: {get_cas}")

    return
