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
from alita.utils.custom_filters import admin_filter, invite_filter, promote_filter
from alita.utils.extract_user import extract_user
from alita.utils.parser import mention_html
from alita.utils.redis_helper import RedisHelper

__PLUGIN__ = "Admin"
__help__ = """
Lazy to promote or demote someone for admins? Want to see basic information about chat? \
All stuff about chatroom such as admin lists, pinning or grabbing an invite link can be \
done easily using the bot.

**User Commands:**
 × /adminlist: List all admins in the current chat.

**Admin only:**
 × /invitelink: Gets private chat's invitelink.
 × /promote: Promotes the user replied to or tagged.
 × /demote: Demotes the user replied to or tagged.

An example of promoting someone to admins:
`/promote @username`; this promotes a user to admin.
"""


@Alita.on_message(filters.command("adminlist", PREFIX_HANDLER) & filters.group)
async def adminlist_show(_, m: Message):

    try:
        try:
            adminlist = (await RedisHelper.get_key("ADMINDICT"))[
                str(m.chat.id)
            ]  # Load ADMINDICT from string
            note = tlang(m, "admin.adminlist.note_cached")
        except Exception:
            adminlist = []
            async for i in m.chat.iter_members(
                filter="administrators",
            ):
                if i.user.is_deleted or i.user.is_bot:
                    continue  # We don't need deleted accounts or bot accounts
                adminlist.append(
                    (
                        i.user.id,
                        f"@{i.user.username}" if i.user.username else i.user.first_name,
                    ),
                )
            adminlist = sorted(adminlist, key=lambda x: x[1])
            note = tlang(m, "admin.adminlist.note_updated")
            ADMINDICT = await RedisHelper.get_key("ADMINDICT")
            ADMINDICT[str(m.chat.id)] = adminlist
            await RedisHelper.set_key("ADMINDICT", ADMINDICT)

        adminstr = (tlang(m, "admin.adminlist.adminstr")).format(
            chat_title=f"<b>{m.chat.title}</b>",
        )

        for i in adminlist:
            try:
                mention = (
                    i[1] if i[1].startswith("@") else (await mention_html(i[1], i[0]))
                )
                adminstr += f"- {mention}\n"
            except PeerIdInvalid:
                pass

        await m.reply_text(f"{adminstr}\n\n<i>Note: {note}</i>")

    except Exception as ef:
        if str(ef) == str(m.chat.id):
            await m.reply_text(tlang(m, "admin.adminlist.use_admin_cache"))
        else:
            ef = str(ef) + f"{adminlist}\n"
            await m.reply_text(
                (tlang(m, "general.some_error")).format(
                    SUPPORT_GROUP=f"@{SUPPORT_GROUP}",
                    ef=f"<code>{ef}</code>",
                ),
            )
            LOGGER.error(ef)

    return


@Alita.on_message(
    filters.command("admincache", PREFIX_HANDLER) & filters.group & admin_filter,
)
async def reload_admins(_, m: Message):

    ADMINDICT = await RedisHelper.get_key("ADMINDICT")  # Load ADMINDICT from string

    try:
        adminlist = []
        async for i in m.chat.iter_members(filter="administrators"):
            if i.user.is_deleted:
                continue
            adminlist.append(
                (
                    i.user.id,
                    f"@{i.user.username}" if i.user.username else i.user.first_name,
                ),
            )
        ADMINDICT[str(m.chat.id)] = adminlist
        await RedisHelper.set_key("ADMINDICT", ADMINDICT)
        await m.reply_text(tlang(m, "admin.adminlist.reloaded_admins"))
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=f"@{SUPPORT_GROUP}",
                ef=f"<code>{ef}</code>",
            ),
        )
        LOGGER.error(ef)
    return


@Alita.on_message(
    filters.command("promote", PREFIX_HANDLER) & filters.group & promote_filter,
)
async def promote_usr(c: Alita, m: Message):

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
            (tlang(m, "admin.promoted_user")).format(
                promoter=(await mention_html(m.from_user.first_name, m.from_user.id)),
                promoted=(await mention_html(user_first_name, user_id)),
                chat_title=f"<b>{m.chat.title}</b>",
            ),
        )

        # ----- Add admin to redis cache! -----
        adminlist = (await RedisHelper.get_key("ADMINDICT"))[
            str(m.chat.id)
        ]  # Load ADMINDICT from string
        u = await m.chat.get_member(user_id)
        adminlist.append(
            [
                u.user.id,
                f"@{u.user.username}" if u.user.username else u.user.first_name,
            ],
        )
        ADMINDICT = await RedisHelper.get_key("ADMINDICT")
        ADMINDICT[str(m.chat.id)] = adminlist
        await RedisHelper.set_key("ADMINDICT", ADMINDICT)

    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except RightForbidden:
        await m.reply_text(tlang(m, "admin.bot_no_promote_right"))
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=f"@{SUPPORT_GROUP}",
                ef=f"<code>{ef}</code>",
            ),
        )
        LOGGER.error(ef)

    return


@Alita.on_message(
    filters.command("demote", PREFIX_HANDLER) & filters.group & promote_filter,
)
async def demote_usr(c: Alita, m: Message):

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
            (tlang(m, "admin.demoted_user")).format(
                demoter=(await mention_html(m.from_user.first_name, m.from_user.id)),
                demoted=(await mention_html(user_first_name, user_id)),
                chat_title=f"<b>{m.chat.title}</b>",
            ),
        )

        # ----- Add admin to redis cache! -----
        ADMINDICT = await RedisHelper.get_key("ADMINDICT")  # Load ADMINDICT from string
        adminlist = []
        async for i in m.chat.iter_members(filter="administrators"):
            if i.user.is_deleted:
                continue
            adminlist.append(
                [
                    i.user.id,
                    f"@{i.user.username}" if i.user.username else i.user.first_name,
                ],
            )
        ADMINDICT[str(m.chat.id)] = adminlist
        await RedisHelper.set_key("ADMINDICT", ADMINDICT)

    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except RightForbidden:
        await m.reply_text(tlang(m, "admin.bot_no_demote_right"))
    except UserAdminInvalid:
        await m.reply_text(tlang(m, "admin.user_admin_invalid"))
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=f"@{SUPPORT_GROUP}",
                ef=f"<code>{ef}</code>",
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
                chat_name=f"<b>{m.chat.id}</b>",
                link=link,
            ),
        )
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except ChatAdminInviteRequired:
        await m.reply_text(tlang(m, "admin.noinviteperm"))
    except RightForbidden:
        await m.reply_text(tlang(m, "no_invite_perm"))
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=f"@{SUPPORT_GROUP}",
                ef=f"<code>{ef}</code>",
            ),
        )
        LOGGER.error(ef)

    return
