from alita.__main__ import Alita
from pyrogram import filters
from pyrogram.types import Message
from alita import (
    WHITELIST_USERS,
    SUDO_USERS,
    DEV_USERS,
    OWNER_ID,
    DEV_PREFIX_HANDLER,
)
from alita.utils.parser import mention_html
from alita.utils.custom_filters import dev_filter


__PLUGIN__ = "Botstaff"


@Alita.on_message(filters.command("botstaff", DEV_PREFIX_HANDLER) & dev_filter)
async def botstaff(c: Alita, m: Message):
    try:
        owner = await c.get_users(OWNER_ID)
        reply = f"<b>🌟 Owner:</b> {mention_html(owner.first_name, OWNER_ID)} (<code>{OWNER_ID}</code>)\n"
    except:
        pass
    true_dev = list(set(DEV_USERS) - {OWNER_ID})
    reply += "\n<b>Developers ⚡️:</b>\n"
    if true_dev == []:
        reply += "No Dev Users"
    else:
        for each_user in true_dev:
            user_id = int(each_user)
            try:
                user = await c.get_users(user_id)
                reply += f"• {mention_html(user.first_name, user_id)} (<code>{user_id}</code>)\n"
            except:
                pass
    true_sudo = list(set(SUDO_USERS) - set(DEV_USERS))
    reply += "\n<b>Sudo Users 🐉:</b>\n"
    if true_sudo == []:
        reply += "No Sudo Users\n"
    else:
        for each_user in true_sudo:
            user_id = int(each_user)
            try:
                user = await c.get_users(user_id)
                reply += f"• {mention_html(user.first_name, user_id)} (<code>{user_id}</code>)\n"
            except:
                pass
    reply += "\n<b>Whitelisted Users 🐺:</b>\n"
    if WHITELIST_USERS == []:
        reply += "No additional whitelisted users\n"
    else:
        for each_user in WHITELIST_USERS:
            user_id = int(each_user)
            try:
                user = await c.get_users(user_id)
                reply += f"• {mention_html(user.first_name, user_id)} (<code>{user_id}</code>)\n"
            except:
                pass
    await m.reply_text(reply)
    return
