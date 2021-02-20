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


from asyncio import sleep
from io import BytesIO
from time import time

from pyrogram import filters
from pyrogram.errors import ChatAdminRequired, RPCError
from pyrogram.types import (
    CallbackQuery,
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    Message,
)

from alita import DEV_PREFIX_HANDLER, LOGGER, PREFIX_HANDLER, SUPPORT_GROUP
from alita.bot_class import Alita
from alita.utils.custom_filters import owner_filter, restrict_filter
from alita.utils.extract_user import extract_user
from alita.utils.localization import GetLang
from alita.utils.parser import mention_html

__PLUGIN__ = "Bans"
__help__ = """
Someone annoying entered your group?
Want to ban/restriction him/her?
This is the plugin for you, easily kick, ban and unban members in a group.

**Admin only:**
 × /kick: Kick the user replied or tagged.
 × /ban: Bans the user replied to or tagged.
 × /unban: Unbans the user replied to or tagged.
 × /banall: Ban all members of a chat!
"""


@Alita.on_message(
    filters.command("kick", PREFIX_HANDLER) & filters.group & restrict_filter,
)
async def kick_usr(_, m: Message):

    _ = GetLang(m).strs

    user_id, user_first_name = await extract_user(m)
    try:
        await m.chat.kick_member(user_id, int(time() + 45))
        await m.reply_text(
            f"Banned {(await mention_html(user_first_name, user_id))}",
        )
    except ChatAdminRequired:
        await m.reply_text(_("admin.notadmin"))
    except RPCError as ef:
        await m.reply_text(f"<code>{ef}</code>\nReport to @{SUPPORT_GROUP}")
        LOGGER.error(ef)

    return


@Alita.on_message(
    filters.command("ban", PREFIX_HANDLER) & filters.group & restrict_filter,
)
async def ban_usr(_, m: Message):

    _ = GetLang(m).strs

    user_id, user_first_name = await extract_user(m)
    try:
        await m.chat.kick_member(user_id)
        await m.reply_text(
            f"Banned {(await mention_html(user_first_name, user_id))}",
        )
    except ChatAdminRequired:
        await m.reply_text(_("admin.notadmin"))
    except RPCError as ef:
        await m.reply_text(f"<code>{ef}</code>\nReport to @{SUPPORT_GROUP}")
        LOGGER.error(ef)

    return


@Alita.on_message(
    filters.command("unban", PREFIX_HANDLER) & filters.group & restrict_filter,
)
async def unban_usr(_, m: Message):

    _ = GetLang(m).strs

    user_id, user_first_name = await extract_user(m)
    try:
        await m.chat.unban_member(user_id)
        await m.reply_text(
            f"Unbanned {(await mention_html(user_first_name, user_id))}",
        )
    except ChatAdminRequired:
        await m.reply_text(_("admin.notadmin"))
    except RPCError as ef:
        await m.reply_text(f"<code>{ef}</code>\nReport to @{SUPPORT_GROUP}")
        LOGGER.error(ef)

    return


@Alita.on_message(filters.command("banall", DEV_PREFIX_HANDLER) & owner_filter)
async def banall_chat(_, m: Message):
    await m.reply_text(
        "Are you sure you want to ban all members in this group?",
        reply_markup=InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton("⚠️ Confirm", callback_data="ban.all.members"),
                    InlineKeyboardButton("❌ Cancel", callback_data="close"),
                ],
            ],
        ),
    )
    return


@Alita.on_callback_query(filters.regex("^ban.all.members$") & owner_filter)
async def banallnotes_callback(_, q: CallbackQuery):

    replymsg = await q.message.edit_text("<i><b>Banning All Members...</b></i>")
    users = []
    fs = 0
    async for x in q.message.chat.iter_members():
        try:
            if fs >= 5:
                await sleep(5)
            await q.message.chat.kick_member(x.user.id)
            users.append(x.user.id)
        except BaseException:
            fs += 1

    rply = f"Users Banned:\n{users}"

    with BytesIO(str.encode(rply)) as f:
        f.name = f"bannedUsers_{q.message.chat.id}.txt"
        await q.message.reply_document(
            document=f,
            caption=f"Banned {len(users)} users!",
        )
        await replymsg.delete()

    await q.answer()
    return
