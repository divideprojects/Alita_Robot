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
    RightForbidden,
    RPCError,
    UserAdminInvalid,
)
from pyrogram.types import Message

from alita import LOGGER, PREFIX_HANDLER, SUPPORT_GROUP, SUPPORT_STAFF
from alita.bot_class import Alita
from alita.database.approve_db import Approve
from alita.tr_engine import tlang
from alita.utils.caching import ADMIN_CACHE, TEMP_ADMIN_CACHE_BLOCK, admin_cache_reload
from alita.utils.custom_filters import admin_filter, invite_filter, promote_filter
from alita.utils.extract_user import extract_user
from alita.utils.parser import mention_html

app_db = Approve()


@Alita.on_message(filters.command("adminlist", PREFIX_HANDLER) & filters.group)
async def adminlist_show(_, m: Message):
    global ADMIN_CACHE
    try:
        try:
            admin_list = ADMIN_CACHE[m.chat.id]
            note = tlang(m, "admin.adminlist.note_cached")
        except KeyError:
            admin_list = await admin_cache_reload(m, "adminlist")
            note = tlang(m, "admin.adminlist.note_updated")

        adminstr = (tlang(m, "admin.adminlist.adminstr")).format(
            chat_title=m.chat.title,
        ) + "\n\n"

        # format is like: (user_id, username/name,anonyamous or not)
        mention = [
            (
                admin[1]
                if admin[1].startswith("@")
                else (await mention_html(admin[1], admin[0]))
            )
            for admin in admin_list
            if not admin[2]  # if non-anonyamous admin
        ]
        mention.sort(key=lambda x: x[1])
        adminstr += "\n".join([f"- {i}" for i in mention])

        await m.reply_text(adminstr + "\n\n" + note)
        LOGGER.info(f"Adminlist cmd use in {m.chat.id} by {m.from_user.id}")

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
    filters.command("admincache", PREFIX_HANDLER) & admin_filter,
)
async def reload_admins(_, m: Message):
    global TEMP_ADMIN_CACHE_BLOCK

    if (m.chat.id in set(TEMP_ADMIN_CACHE_BLOCK.keys())) and (
        m.from_user.id not in SUPPORT_STAFF
    ):
        if TEMP_ADMIN_CACHE_BLOCK[m.chat.id] == "manualblock":
            await m.reply_text("Can only reload admin cache once per 10 mins!")
            return

    try:
        await admin_cache_reload(m, "admincache")
        TEMP_ADMIN_CACHE_BLOCK[m.chat.id] = "manualblock"
        await m.reply_text(tlang(m, "admin.adminlist.reloaded_admins"))
        LOGGER.info(f"Admincache cmd use in {m.chat.id} by {m.from_user.id}")
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
    filters.command("promote", PREFIX_HANDLER) & promote_filter,
)
async def promote_usr(c: Alita, m: Message):
    from alita import BOT_ID

    global ADMIN_CACHE

    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text(tlang(m, "admin.promote.no_target"))
        return

    user_id, user_first_name, user_name = await extract_user(c, m)

    if user_id == BOT_ID:
        await m.reply_text("Huh, how can I even promote myself?")
        return

    # If user is alreay admin
    try:
        admin_list = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admin_list = {
            i[0] for i in (await admin_cache_reload(m, "promote_cache_update"))
        }

    if user_id in admin_list:
        await m.reply_text(
            "This user is already an admin, how am I supposed to re-promote them?",
        )
        return

    try:
        await m.chat.promote_member(
            user_id=user_id,
            can_change_info=False,
            can_delete_messages=True,
            can_restrict_members=True,
            can_invite_users=True,
            can_pin_messages=True,
            can_manage_voice_chats=False,
        )
        LOGGER.info(f"{m.from_user.id} promoted {user_id} in {m.chat.id}")

        await m.reply_text(
            (tlang(m, "admin.promote.promoted_user")).format(
                promoter=(await mention_html(m.from_user.first_name, m.from_user.id)),
                promoted=(await mention_html(user_first_name, user_id)),
                chat_title=m.chat.title,
            ),
        )

        if len(m.text.split()) == 3 and not m.reply_to_message:
            await c.set_administrator_title(m.chat.id, user_id, m.text.split()[2])
        elif len(m.text.split()) == 2 and m.reply_to_message:
            await c.set_administrator_title(m.chat.id, user_id, m.text.split()[1])

        # If user is approved, disapprove them as they willbe promoted and get even more rights
        if app_db.check_approve(m.chat.id, user_id):
            app_db.remove_approve(m.chat.id, user_id)

        # ----- Add admin to temp cache -----
        try:
            inp1 = user_name if user_name else user_first_name
            admins_group = ADMIN_CACHE[m.chat.id]
            admins_group.append((user_id, inp1))
            ADMIN_CACHE[m.chat.id] = admins_group
        except KeyError:
            await admin_cache_reload(m, "promote_key_error")

    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except RightForbidden:
        await m.reply_text(tlang(m, "admin.promote.bot_no_right"))
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
        LOGGER.error(format_exc())

    return


@Alita.on_message(
    filters.command("demote", PREFIX_HANDLER) & promote_filter,
)
async def demote_usr(c: Alita, m: Message):
    from alita import BOT_ID

    global ADMIN_CACHE

    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text(tlang(m, "admin.demote.no_target"))
        return

    user_id, user_first_name, _ = await extract_user(c, m)

    if user_id == BOT_ID:
        await m.reply_text("Get an admin to demote me!")
        return

    # If user not alreay admin
    try:
        admin_list = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admin_list = {
            i[0] for i in (await admin_cache_reload(m, "demote_cache_update"))
        }

    if user_id not in admin_list:
        await m.reply_text(
            "This user is not an admin, how am I supposed to re-demote them?",
        )
        return

    try:
        await m.chat.promote_member(
            user_id=user_id,
            can_change_info=False,
            can_delete_messages=False,
            can_restrict_members=False,
            can_invite_users=False,
            can_pin_messages=False,
            can_manage_voice_chats=False,
        )
        LOGGER.info(f"{m.from_user.id} demoted {user_id} in {m.chat.id}")

        # ----- Remove admin from cache -----
        try:
            admin_list = ADMIN_CACHE[m.chat.id]
            user = next(user for user in admin_list if user[0] == user_id)
            admin_list.remove(user)
            ADMIN_CACHE[m.chat.id] = admin_list
        except (KeyError, StopIteration):
            await admin_cache_reload(m, "demote_key_stopiter_error")

        await m.reply_text(
            (tlang(m, "admin.demote.demoted_user")).format(
                demoter=(await mention_html(m.from_user.first_name, m.from_user.id)),
                demoted=(await mention_html(user_first_name, user_id)),
                chat_title=m.chat.title,
            ),
        )

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
        LOGGER.error(format_exc())

    return


@Alita.on_message(
    filters.command("invitelink", PREFIX_HANDLER) & invite_filter,
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
        LOGGER.info(f"{m.from_user.id} exported invite link in {m.chat.id}")
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
        LOGGER.error(format_exc())

    return


__PLUGIN__ = "plugins.admin.main"
__help__ = "plugins.admin.help"
