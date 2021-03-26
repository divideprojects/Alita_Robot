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


from time import time
from traceback import format_exc

from pyrogram import filters
from pyrogram.errors import (
    ChatAdminRequired,
    PeerIdInvalid,
    RightForbidden,
    RPCError,
    UserAdminInvalid,
)
from pyrogram.types import Message

from alita import LOGGER, PREFIX_HANDLER, SUPPORT_GROUP, SUPPORT_STAFF
from alita.bot_class import Alita
from alita.tr_engine import tlang
from alita.utils.caching import ADMIN_CACHE, admin_cache_reload
from alita.utils.custom_filters import restrict_filter
from alita.utils.extract_user import extract_user
from alita.utils.parser import mention_html


@Alita.on_message(
    filters.command(["kick", "skick", "dkick"], PREFIX_HANDLER) & restrict_filter,
)
async def kick_usr(c: Alita, m: Message):
    from alita import BOT_ID

    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text(tlang(m, "admin.kick.no_target"))
        await m.stop_propagation()

    if m.reply_to_message and len(m.text.split()) >= 2:
        reason = m.text.split(None, 1)[1]
    elif not m.reply_to_message and len(m.text.split()) >= 3:
        reason = m.text.split(None, 2)[2]
    else:
        reason = None

    user_id, user_first_name, _ = await extract_user(c, m)

    if user_id == BOT_ID:
        await m.reply_text("Huh, why would I kick myself?")
        await m.stop_propagation()

    if user_id in SUPPORT_STAFF:
        await m.reply_text(tlang(m, "admin.support_cannot_restrict"))
        LOGGER.info(
            f"{m.from_user.id} trying to kick {user_id} (SUPPORT_STAFF) in {m.chat.id}",
        )
        await m.stop_propagation()

    try:
        admins_group = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admins_group = await admin_cache_reload(m, "kick")

    if user_id in admins_group:
        await m.reply_text(tlang(m, "admin.kick.admin_cannot_kick"))
        await m.stop_propagation()

    try:
        LOGGER.info(f"{m.from_user.id} kicked {user_id} in {m.chat.id}")
        await c.kick_chat_member(m.chat.id, user_id, int(time() + 45))
        if m.text.split()[0] == "/skick":
            await m.delete()
            await m.stop_propagation()
        if m.text.split()[0] == "/dkick":
            if not m.reply_to_message:
                await m.reply_text("Reply to a message to delete it and kick the user!")
                await m.stop_propagation()
            await m.reply_to_message.delete()
        txt = (tlang(m, "admin.kick.kicked_user")).format(
            admin=(await mention_html(m.from_user.first_name, m.from_user.id)),
            kicked=(await mention_html(user_first_name, user_id)),
            chat_title=m.chat.title,
        )
        txt += f"\n<b>Reason</b>: {reason}" if reason else ""
        await m.reply_text(txt)
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except PeerIdInvalid:
        await m.reply_text(
            "I have not seen this user yet...!\nMind forwarding one of their message so I can recognize them?",
        )
    except UserAdminInvalid:
        await m.reply_text(tlang(m, "admin.user_admin_invalid"))
    except RightForbidden:
        await m.reply_text(tlang(m, "admin.kick.bot_no_right"))
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)
        LOGGER.error(format_exc())

    await m.stop_propagation()


@Alita.on_message(
    filters.command(["ban", "sban", "dban"], PREFIX_HANDLER) & restrict_filter,
)
async def ban_usr(c: Alita, m: Message):
    from alita import BOT_ID

    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text(tlang(m, "admin.ban.no_target"))
        await m.stop_propagation()

    if m.reply_to_message and len(m.text.split()) >= 2:
        reason = m.text.split(None, 1)[1]
    elif not m.reply_to_message and len(m.text.split()) >= 3:
        reason = m.text.split(None, 2)[2]
    else:
        reason = None

    user_id, user_first_name, _ = await extract_user(c, m)

    if user_id == BOT_ID:
        await m.reply_text("Huh, why would I ban myself?")
        await m.stop_propagation()

    if user_id in SUPPORT_STAFF:
        await m.reply_text(tlang(m, "admin.support_cannot_restrict"))
        LOGGER.info(
            f"{m.from_user.id} trying to ban {user_id} (SUPPORT_STAFF) in {m.chat.id}",
        )
        await m.stop_propagation()

    try:
        admins_group = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admins_group = await admin_cache_reload(m, "ban")

    if user_id in admins_group:
        await m.reply_text(tlang(m, "admin.ban.admin_cannot_ban"))
        await m.stop_propagation()

    try:
        LOGGER.info(f"{m.from_user.id} banned {user_id} in {m.chat.id}")
        await c.kick_chat_member(m.chat.id, user_id)
        if m.text.split()[0] == "/sban":
            await m.delete()
            await m.stop_propagation()
        if m.text.split()[0] == "/dban":
            if not m.reply_to_message:
                await m.reply_text("Reply to a message to delete it and ban the user!")
                await m.stop_propagation()
            await m.reply_to_message.delete()
        txt = (tlang(m, "admin.ban.banned_user")).format(
            admin=(await mention_html(m.from_user.first_name, m.from_user.id)),
            banned=(await mention_html(user_first_name, user_id)),
            chat_title=m.chat.title,
        )
        txt += f"\n<b>Reason</b>: {reason}" if reason else ""
        await m.reply_text(txt)
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except PeerIdInvalid:
        await m.reply_text(
            "I have not seen this user yet...!\nMind forwarding one of their message so I can recognize them?",
        )
    except UserAdminInvalid:
        await m.reply_text(tlang(m, "admin.user_admin_invalid"))
    except RightForbidden:
        await m.reply_text(tlang(m, tlang(m, "admin.ban.bot_no_right")))
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)
        LOGGER.error(format_exc())
    await m.stop_propagation()


@Alita.on_message(filters.command("unban", PREFIX_HANDLER) & restrict_filter)
async def unban_usr(c: Alita, m: Message):

    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text(tlang(m, "admin.unban.no_target"))
        await m.stop_propagation()

    user_id, user_first_name, _ = await extract_user(c, m)

    if m.reply_to_message and len(m.text.split()) >= 2:
        reason = m.text.split(None, 1)[1]
    elif not m.reply_to_message and len(m.text.split()) >= 3:
        reason = m.text.split(None, 2)[2]
    else:
        reason = None

    try:
        await m.chat.unban_member(user_id)
        txt = (tlang(m, "admin.unban.unbanned_user")).format(
            admin=(await mention_html(m.from_user.first_name, m.from_user.id)),
            unbanned=(await mention_html(user_first_name, user_id)),
            chat_title=m.chat.title,
        )
        txt += f"\n<b>Reason</b>: {reason}" if reason else ""
        await m.reply_text(txt)
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
        LOGGER.error(format_exc())

    await m.stop_propagation()


__PLUGIN__ = "plugins.bans.main"
__help__ = "plugins.bans.help"
__alt_name__ = ["ban", "unban", "kick"]
