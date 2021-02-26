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
from pyrogram.errors import ChatAdminRequired, RightForbidden, RPCError
from pyrogram.types import (
    CallbackQuery,
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    Message,
)

from alita import (
    DEV_PREFIX_HANDLER,
    LOGGER,
    PREFIX_HANDLER,
    SUPPORT_GROUP,
    SUPPORT_STAFF,
)
from alita.bot_class import Alita
from alita.tr_engine import tlang
from alita.utils.custom_filters import owner_filter, restrict_filter
from alita.utils.extract_user import extract_user
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

    user_id, user_first_name = await extract_user(m)

    if user_id in SUPPORT_STAFF:
        await m.reply_text("This user is in my support staff, cannot restrict them.")
        return

    try:
        await m.chat.kick_member(user_id, int(time() + 45))
        await m.reply_text(
            (await tlang(m, "admin.kicked_user")).format(
                admin=(await mention_html(m.from_user.first_name, m.from_user.id)),
                kicked=(await mention_html(user_first_name, user_id)),
                chat_title=f"<b>{m.chat.title}</b>",
            ),
        )
    except ChatAdminRequired:
        await m.reply_text(await tlang(m, "admin.not_admin"))
    except RightForbidden:
        await m.reply_text(await tlang(m, "admin.bot_no_kick_right"))
    except RPCError as ef:
        await m.reply_text(
            (await tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=f"@{SUPPORT_GROUP}",
                ef=f"<code>{ef}</code>",
            ),
        )
        LOGGER.error(ef)

    return


@Alita.on_message(
    filters.command("ban", PREFIX_HANDLER) & filters.group & restrict_filter,
)
async def ban_usr(_, m: Message):

    user_id, user_first_name = await extract_user(m)

    if user_id in SUPPORT_STAFF:
        await m.reply_text("This user is in my support staff, cannot restrict them.")
        return

    try:
        await m.chat.kick_member(user_id)
        await m.reply_text(
            (await tlang(m, "admin.banned_user")).format(
                admin=(await mention_html(m.from_user.first_name, m.from_user.id)),
                banned=(await mention_html(user_first_name, user_id)),
                chat_title=f"<b>{m.chat.title}</b>",
            ),
        )
    except ChatAdminRequired:
        await m.reply_text(await tlang(m, "admin.not_admin"))
    except RightForbidden:
        await m.reply_text(await tlang(m, await tlang(m, "admin.bot_no_ban_right")))
    except RPCError as ef:
        await m.reply_text(
            (await tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=f"@{SUPPORT_GROUP}",
                ef=f"<code>{ef}</code>",
            ),
        )
        LOGGER.error(ef)

    return


@Alita.on_message(
    filters.command("unban", PREFIX_HANDLER) & filters.group & restrict_filter,
)
async def unban_usr(_, m: Message):

    user_id, user_first_name = await extract_user(m)

    if user_id in SUPPORT_STAFF:
        await m.reply_text("This user is in my support staff, cannot restrict them.")
        return

    try:
        await m.chat.unban_member(user_id)
        await m.reply_text(
            (await tlang(m, "admin.banned_user")).format(
                admin=(await mention_html(m.from_user.first_name, m.from_user.id)),
                unbanned=(await mention_html(user_first_name, user_id)),
                chat_title=f"<b>{m.chat.title}</b>",
            ),
        )
    except ChatAdminRequired:
        await m.reply_text(await tlang(m, "admin.not_admin"))
    except RightForbidden:
        await m.reply_text(await tlang(m, await tlang(m, "admin.bot_no_unban_right")))
    except RPCError as ef:
        await m.reply_text(
            (await tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=f"@{SUPPORT_GROUP}",
                ef=f"<code>{ef}</code>",
            ),
        )
        LOGGER.error(ef)

    return


@Alita.on_message(filters.command("banall", DEV_PREFIX_HANDLER) & owner_filter)
async def banall_chat(_, m: Message):
    await m.reply_text(
        (await tlang(m, "admin.ban_all")),
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

    replymsg = await q.message.edit_text(
        f"<i><b>{(await tlang(q, 'admin.banning_all'))}</b></i>",
    )
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
