# Copyright (C) 2020 - 2021 Divkix. All rights reserved. Source code available under the AGPL.
#
# This file is part of Alita_Robot.
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as
# published by the Free Software Foundation, either version 3 of the
# License, or (at your option) any later version.

# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.

# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.


from pyrogram import filters
from pyrogram.types import Message
from alita import (
    WHITELIST_USERS,
    SUDO_USERS,
    DEV_USERS,
    OWNER_ID,
    DEV_PREFIX_HANDLER,
)
from alita.bot_class import Alita
from alita.utils.parser import mention_html
from alita.utils.custom_filters import dev_filter


__PLUGIN__ = "Botstaff"


@Alita.on_message(filters.command("botstaff", DEV_PREFIX_HANDLER) & dev_filter)
async def botstaff(c: Alita, m: Message):
    try:
        owner = await c.get_users(OWNER_ID)
        reply = f"<b>🌟 Owner:</b> {(await mention_html(owner.first_name, OWNER_ID))} (<code>{OWNER_ID}</code>)\n"
    except BaseException:
        pass
    true_dev = list(set(DEV_USERS) - {OWNER_ID})
    reply += "\n<b>Developers ⚡️:</b>\n"
    if true_dev == []:
        reply += "No Dev Users\n"
    else:
        for each_user in true_dev:
            user_id = int(each_user)
            try:
                user = await c.get_users(user_id)
                reply += f"• {(await mention_html(user.first_name, user_id))} (<code>{user_id}</code>)\n"
            except BaseException:
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
                reply += f"• {(await mention_html(user.first_name, user_id))} (<code>{user_id}</code>)\n"
            except BaseException:
                pass
    reply += "\n<b>Whitelisted Users 🐺:</b>\n"
    if WHITELIST_USERS == []:
        reply += "No additional whitelisted users\n"
    else:
        for each_user in WHITELIST_USERS:
            user_id = int(each_user)
            try:
                user = await c.get_users(user_id)
                reply += f"• {(await mention_html(user.first_name, user_id))} (<code>{user_id}</code>)\n"
            except BaseException:
                pass
    await m.reply_text(reply)
    return
