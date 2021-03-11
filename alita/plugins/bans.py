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
from traceback import format_exc

from pyrogram import filters
from pyrogram.errors import (
    ChatAdminRequired,
    RightForbidden,
    RPCError,
    UserNotParticipant,
)
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
from alita.utils.admin_cache import ADMIN_CACHE
from alita.utils.clean_file import remove_markdown_and_html
from alita.utils.custom_filters import owner_filter, restrict_filter
from alita.utils.extract_user import extract_user
from alita.utils.parser import mention_html

__PLUGIN__ = "plugins.bans.main"
__help__ = "plugins.bans.help"


@Alita.on_message(
    filters.command("kick", PREFIX_HANDLER) & filters.group & restrict_filter,
)
async def kick_usr(c: Alita, m: Message):

    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text(tlang(m, "admin.kick.no_target"))
        return

    user_id, user_first_name = await extract_user(c, m)

    if user_id in SUPPORT_STAFF:
        await m.reply_text(tlang(m, "admin.support_cannot_restrict"))
        return

    if user_id in [i[0] for i in ADMIN_CACHE[str(m.chat.id)]]:
        await m.reply_text(tlang(m, "admin.kick.admin_cannot_kick"))
        return

    try:
        # Check if user is banned or not
        # banned_users = []
        # async for i in m.chat.iter_members(filter="kicked"):
        #     banned_users.append(i.user.id)
        # if user_id in banned_users:
        #     await m.reply_text(tlang(m, "admin.kick.user_already_banned"))
        #     return
        await m.chat.kick_member(user_id, int(time() + 45))
        await m.reply_text(
            (tlang(m, "admin.kick.kicked_user")).format(
                admin=(await mention_html(m.from_user.first_name, m.from_user.id)),
                kicked=(await mention_html(user_first_name, user_id)),
                chat_title=m.chat.title,
            ),
        )
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except RightForbidden:
        await m.reply_text(tlang(m, "admin.kick.bot_no_right"))
    except UserNotParticipant:
        await m.reply_text("How can I kick a member who is not in this chat?")
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)

    return


@Alita.on_message(
    filters.command("ban", PREFIX_HANDLER) & filters.group & restrict_filter,
)
async def ban_usr(c: Alita, m: Message):

    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text(tlang(m, "admin.ban.no_target"))
        return

    user_id, user_first_name = await extract_user(c, m)

    if user_id in SUPPORT_STAFF:
        await m.reply_text(tlang(m, "admin.support_cannot_restrict"))
        return

    if user_id in [i[0] for i in ADMIN_CACHE[str(m.chat.id)]]:
        await m.reply_text(tlang(m, "admin.kick.admin_cannot_kick"))
        return

    try:
        # Check if user is banned or not
        # banned_users = []
        # async for i in m.chat.iter_members(filter="kicked"):
        #     banned_users.append(i.user.id)
        # if user_id in banned_users:
        #     await m.reply_text(tlang(m, "admin.kick.user_already_banned"))
        #     return
        await m.chat.kick_member(user_id)
        await m.reply_text(
            (tlang(m, "admin.ban.banned_user")).format(
                admin=(await mention_html(m.from_user.first_name, m.from_user.id)),
                banned=(await mention_html(user_first_name, user_id)),
                chat_title=m.chat.title,
            ),
        )
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except RightForbidden:
        await m.reply_text(tlang(m, tlang(m, "admin.ban.bot_no_right")))
    except UserNotParticipant:
        await m.reply_text("How can I ban a member who is not in this chat?")
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)

    return


@Alita.on_message(
    filters.command("unban", PREFIX_HANDLER) & filters.group & restrict_filter,
)
async def unban_usr(c: Alita, m: Message):

    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text(tlang(m, "admin.unban.no_target"))
        return

    user_id, user_first_name = await extract_user(c, m)

    try:

        # Check if user is banned or not
        banned_users = []
        async for i in m.chat.iter_members(filter="kicked"):
            banned_users.append(i.user.id)
        if user_id not in banned_users:
            await m.reply_text(tlang(m, "admin.unban.user_not_banned"))
            return

        await m.chat.unban_member(user_id)
        await m.reply_text(
            (tlang(m, "admin.unban.unbanned_user")).format(
                admin=(await mention_html(m.from_user.first_name, m.from_user.id)),
                unbanned=(await mention_html(user_first_name, user_id)),
                chat_title=m.chat.title,
            ),
        )
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except RightForbidden:
        await m.reply_text(tlang(m, tlang(m, "admin.unban.bot_no_right")))
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)

    return


@Alita.on_message(filters.command("banall", DEV_PREFIX_HANDLER) & owner_filter)
async def banall_chat(_, m: Message):
    await m.reply_text(
        (tlang(m, "admin.ban.ban_all")),
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
        f"<i><b>{(tlang(q, 'admin.ban.banning_all'))}</b></i>",
    )
    users = []
    fs = 0
    async for x in q.message.chat.iter_members():
        try:
            if fs >= 5:
                await sleep(5)
            await q.message.chat.kick_member(x.user.id)
            users.append(x.user.id)
        except Exception:
            fs += 1
            LOGGER.error(format_exc())

    rply = f"Users Banned:\n{users}"

    with BytesIO(str.encode(remove_markdown_and_html(rply))) as f:
        f.name = f"bannedUsers_{q.message.chat.id}.txt"
        await q.message.reply_document(
            document=f,
            caption=f"Banned {len(users)} users!",
        )
        await replymsg.delete()

    await q.answer()
    return
