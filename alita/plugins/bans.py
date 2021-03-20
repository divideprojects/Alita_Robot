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
from time import time
from traceback import format_exc

from pyrogram import filters
from pyrogram.errors import ChatAdminRequired, RightForbidden, RPCError
from pyrogram.types import Message

from alita import LOGGER, PREFIX_HANDLER, SUPPORT_GROUP, SUPPORT_STAFF
from alita.bot_class import Alita
from alita.tr_engine import tlang
from alita.utils.admin_cache import ADMIN_CACHE, admin_cache_reload
from alita.utils.custom_filters import restrict_filter
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

    user_id, user_first_name, _ = await extract_user(c, m)

    if user_id in SUPPORT_STAFF:
        await m.reply_text(tlang(m, "admin.support_cannot_restrict"))
        LOGGER.info(
            f"{m.from_user.id} trying to kick {user_id} (SUPPORT_STAFF) in {m.chat.id}",
        )
        return

    try:
        admins_group = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admins_group = await admin_cache_reload(m, "kick")

    if user_id in admins_group:
        await m.reply_text(tlang(m, "admin.kick.admin_cannot_kick"))
        return

    try:
        await c.kick_chat_member(m.chat.id, user_id, int(time() + 45))
        LOGGER.info(f"{m.from_user.id} kicked {user_id} in {m.chat.id}")
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
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)
        LOGGER.error(format_exc())

    return


@Alita.on_message(
    filters.command("skick", PREFIX_HANDLER) & filters.group & restrict_filter,
)
async def skick_usr(c: Alita, m: Message):

    if len(m.text.split()) == 1 and not m.reply_to_message:
        mymsg = await m.reply_text(tlang(m, "admin.kick.no_target"))
        sleep(3)
        await m.delete()
        await mymsg.delete()
        return

    user_id = (await extract_user(c, m))[0]

    if user_id in SUPPORT_STAFF:
        mymsg = await m.reply_text(tlang(m, "admin.support_cannot_restrict"))
        sleep(3)
        await m.delete()
        await mymsg.delete()
        LOGGER.info(
            f"{m.from_user.id} trying to skick {user_id} (SUPPORT_STAFF) in {m.chat.id}",
        )
        return

    try:
        admins_group = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admins_group = await admin_cache_reload(m, "skick")

    if user_id in admins_group:
        mymsg = await m.reply_text(tlang(m, "admin.kick.admin_cannot_kick"))
        sleep(3)
        await m.delete()
        await mymsg.delete()
        return

    try:
        await c.kick_chat_member(m.chat.id, user_id, int(time() + 15))
        LOGGER.info(f"{m.from_user.id} skicked {user_id} in {m.chat.id}")
        await m.delete()
    except ChatAdminRequired:
        mymsg = await m.reply_text(tlang(m, "admin.not_admin"))
        sleep(3)
        await m.delete()
        await mymsg.delete()
    except RightForbidden:
        mymsg = await m.reply_text(tlang(m, "admin.kick.bot_no_right"))
        sleep(3)
        await m.delete()
        await mymsg.delete()
    except RPCError as ef:
        mymsg = await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        sleep(3)
        await m.delete()
        await mymsg.delete()
        LOGGER.error(ef)
        LOGGER.error(format_exc())
    return


@Alita.on_message(
    filters.command("ban", PREFIX_HANDLER) & filters.group & restrict_filter,
)
async def ban_usr(c: Alita, m: Message):

    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text(tlang(m, "admin.ban.no_target"))
        return

    user_id, user_first_name, _ = await extract_user(c, m)

    if user_id in SUPPORT_STAFF:
        await m.reply_text(tlang(m, "admin.support_cannot_restrict"))
        LOGGER.info(
            f"{m.from_user.id} trying to ban {user_id} (SUPPORT_STAFF) in {m.chat.id}",
        )
        return

    try:
        admins_group = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admins_group = await admin_cache_reload(m, "ban")

    if user_id in admins_group:
        await m.reply_text(tlang(m, "admin.ban.admin_cannot_ban"))
        return

    try:
        await c.kick_chat_member(m.chat.id, user_id)
        LOGGER.info(f"{m.from_user.id} banned {user_id} in {m.chat.id}")
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
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)
        LOGGER.error(format_exc())
    return


@Alita.on_message(
    filters.command("sban", PREFIX_HANDLER) & filters.group & restrict_filter,
)
async def sban_usr(c: Alita, m: Message):

    if len(m.text.split()) == 1 and not m.reply_to_message:
        mymsg = await m.reply_text(tlang(m, "admin.ban.no_target"))
        sleep(3)
        await m.delete()
        await mymsg.delete()
        return

    user_id = (await extract_user(c, m))[0]

    if user_id in SUPPORT_STAFF:
        mymsg = await m.reply_text(tlang(m, "admin.support_cannot_restrict"))
        sleep(3)
        await m.delete()
        await mymsg.delete()
        LOGGER.info(
            f"{m.from_user.id} trying to sban {user_id} (SUPPORT_STAFF) in {m.chat.id}",
        )
        return

    try:
        admins_group = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admins_group = await admin_cache_reload(m, "sban")

    if user_id in admins_group:
        mymsg = await m.reply_text(tlang(m, "admin.kick.admin_cannot_kick"))
        sleep(3)
        await m.delete()
        await mymsg.delete()
        return

    try:
        await c.kick_chat_member(m.chat.id, user_id)
        LOGGER.info(f"{m.from_user.id} sbanned {user_id} in {m.chat.id}")
        await m.delete()
    except ChatAdminRequired:
        mymsg = await m.reply_text(tlang(m, "admin.not_admin"))
        sleep(3)
        await m.delete()
        await mymsg.delete()
    except RightForbidden:
        mymsg = await m.reply_text(tlang(m, tlang(m, "admin.ban.bot_no_right")))
        sleep(3)
        await m.delete()
        await mymsg.delete()
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)
        LOGGER.error(format_exc())
    return


@Alita.on_message(
    filters.command("unban", PREFIX_HANDLER) & filters.group & restrict_filter,
)
async def unban_usr(c: Alita, m: Message):

    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text(tlang(m, "admin.unban.no_target"))
        return

    user_id, user_first_name, _ = await extract_user(c, m)

    try:
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
        LOGGER.error(format_exc())

    return
