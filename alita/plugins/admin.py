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


from traceback import format_exc

from pyrogram import filters
from pyrogram.errors import (
    ChatAdminInviteRequired,
    ChatAdminRequired,
    PeerIdInvalid,
    RightForbidden,
    RPCError,
    UserAdminInvalid,
)
from pyrogram.types import Message

from alita import LOGGER, PREFIX_HANDLER, SUPPORT_GROUP
from alita.bot_class import Alita
from alita.tr_engine import tlang
from alita.utils.admin_cache import ADMIN_CACHE, admin_cache_reload
from alita.utils.custom_filters import admin_filter, invite_filter, promote_filter
from alita.utils.extract_user import extract_user
from alita.utils.parser import mention_html

__PLUGIN__ = "plugins.admin.main"
__help__ = "plugins.admin.help"


@Alita.on_message(filters.command("adminlist", PREFIX_HANDLER) & filters.group)
async def adminlist_show(_, m: Message):
    global ADMIN_CACHE
    try:
        try:
            admin_list = ADMIN_CACHE[str(m.chat.id)]
            note = tlang(m, "admin.adminlist.note_cached")
        except KeyError:
            admin_list = []
            async for i in m.chat.iter_members(
                filter="administrators",
            ):
                if i.user.is_deleted or i.user.is_bot:
                    continue  # We don't need deleted accounts or bot accounts
                admin_list.append(
                    (
                        i.user.id,
                        ("@" + i.user.username)
                        if i.user.username
                        else i.user.first_name,
                    ),
                )
            admin_list = sorted(admin_list, key=lambda x: x[1])
            note = tlang(m, "admin.adminlist.note_updated")
            ADMIN_CACHE[str(m.chat.id)] = admin_list

        adminstr = (tlang(m, "admin.adminlist.adminstr")).format(
            chat_title=m.chat.title,
        )

        for i in admin_list:
            try:
                mention = (
                    i[1] if i[1].startswith("@") else (await mention_html(i[1], i[0]))
                )
                adminstr += f"- {mention}\n"
            except PeerIdInvalid:
                pass

        await m.reply_text(adminstr + "\n" + note)

    except Exception as ef:
        if str(ef) == str(m.chat.id):
            await m.reply_text(tlang(m, "admin.adminlist.use_admin_cache"))
        else:
            ef = str(ef) + f"{admin_list}\n"
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
    filters.command("admincache", PREFIX_HANDLER) & filters.group & admin_filter,
)
async def reload_admins(_, m: Message):
    try:
        await admin_cache_reload(m)
        await m.reply_text(tlang(m, "admin.adminlist.reloaded_admins"))
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
    filters.command("promote", PREFIX_HANDLER) & filters.group & promote_filter,
)
async def promote_usr(c: Alita, m: Message):

    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text(tlang(m, "admin.promote.no_target"))
        return

    user_id, user_first_name = await extract_user(c, m)
    try:
        await m.chat.promote_member(
            user_id=user_id,
            can_change_info=False,
            can_delete_messages=True,
            can_restrict_members=True,
            can_invite_users=True,
            can_pin_messages=True,
        )
        await m.reply_text(
            (tlang(m, "admin.promote.promoted_user")).format(
                promoter=(await mention_html(m.from_user.first_name, m.from_user.id)),
                promoted=(await mention_html(user_first_name, user_id)),
                chat_title=m.chat.title,
            ),
        )

        # ----- Add admin to temp cache -----
        try:
            global ADMIN_CACHE
            admin_list = ADMIN_CACHE[str(m.chat.id)]  # Load Admins from cached list
        except KeyError:
            admin_list = []
            async for i in m.chat.iter_members(filter="administrators"):
                if (
                    i.user.is_deleted or i.user.is_bot
                ):  # Don't cache deleted users and bots!
                    continue
                admin_list.append(
                    (
                        i.user.id,
                        ("@" + i.user.username)
                        if i.user.username
                        else i.user.first_name,
                    ),
                )

        u = await m.chat.get_member(user_id)
        admin_list.append(
            [
                u.user.id,
                ("@" + u.user.username) if u.user.username else u.user.first_name,
            ],
        )
        admin_list = admin_list = sorted(admin_list, key=lambda x: x[1])
        ADMIN_CACHE[str(m.chat.id)] = admin_list

    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except RightForbidden:
        await m.reply_text(tlang(m, "admin.promote.bot_no_right"))
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
    filters.command("demote", PREFIX_HANDLER) & filters.group & promote_filter,
)
async def demote_usr(c: Alita, m: Message):

    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text(tlang(m, "admin.demote.no_target"))
        return

    user_id, user_first_name = await extract_user(c, m)
    try:
        await m.chat.promote_member(
            user_id=user_id,
            can_change_info=False,
            can_delete_messages=False,
            can_restrict_members=False,
            can_invite_users=False,
            can_pin_messages=False,
        )
        await m.reply_text(
            (tlang(m, "admin.demote.demoted_user")).format(
                demoter=(await mention_html(m.from_user.first_name, m.from_user.id)),
                demoted=(await mention_html(user_first_name, user_id)),
                chat_title=m.chat.title,
            ),
        )

        # ----- Remove admin from cache -----
        try:
            global ADMIN_CACHE
            admin_list = ADMIN_CACHE[str(m.chat.id)]
            user = next(user for user in admin_list if user[0] == user_id)
            admin_list.remove(user)
        except (KeyError, StopIteration):
            admin_list = []
            async for i in m.chat.iter_members(filter="administrators"):
                if i.user.is_deleted or i.user.is_bot:
                    continue
                admin_list.append(
                    [
                        i.user.id,
                        ("@" + i.user.username)
                        if i.user.username
                        else i.user.first_name,
                    ],
                )
            admin_list = admin_list = sorted(admin_list, key=lambda x: x[1])
            ADMIN_CACHE[str(m.chat.id)] = admin_list

    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except RightForbidden:
        await m.reply_text(tlang(m, "admin.demote.bot_no_right"))
    except UserAdminInvalid:
        await m.reply_text(tlang(m, "admin.user_admin_invalid"))
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
    filters.command("invitelink", PREFIX_HANDLER) & filters.group & invite_filter,
)
async def get_invitelink(c: Alita, m: Message):

    try:
        link = await c.export_chat_invite_link(m.chat.id)
        await m.reply_text(
            (tlang(m, "admin.invitelink")).format(
                chat_name=m.chat.id,
                link=link,
            ),
            disable_web_page_preview=True,
        )
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except ChatAdminInviteRequired:
        await m.reply_text(tlang(m, "admin.no_invite_perm"))
    except RightForbidden:
        await m.reply_text(tlang(m, "admin.no_user_invite_perm"))
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)

    return
