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

from pyrogram.errors import (
    ChatAdminRequired,
    RightForbidden,
    RPCError,
    UserNotParticipant,
)
from pyrogram.types import ChatPermissions, Message

from alita import BOT_ID, LOGGER, SUPPORT_GROUP, SUPPORT_STAFF
from alita.bot_class import Alita
from alita.tr_engine import tlang
from alita.utils.caching import ADMIN_CACHE, admin_cache_reload
from alita.utils.custom_filters import command, restrict_filter
from alita.utils.extract_user import extract_user
from alita.utils.parser import mention_html
from alita.utils.string import extract_time


@Alita.on_message(command("tmute") & restrict_filter)
async def tmute_usr(c: Alita, m: Message):
    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text("I can't mute nothing!")
        return

    try:
        user_id, user_first_name, _ = await extract_user(c, m)
    except Exception:
        return

    if not user_id:
        await m.reply_text("Cannot find user to mute !")
        return
    if user_id == BOT_ID:
        await m.reply_text("Huh, why would I mute myself?")
        return

    if user_id in SUPPORT_STAFF:
        LOGGER.info(
            f"{m.from_user.id} trying to mute {user_id} (SUPPORT_STAFF) in {m.chat.id}",
        )
        await m.reply_text(tlang(m, "admin.support_cannot_restrict"))
        return

    try:
        admins_group = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admins_group = await admin_cache_reload(m, "mute")

    if user_id in admins_group:
        await m.reply_text(tlang(m, "admin.mute.admin_cannot_mute"))
        return

    if m.reply_to_message:
        r_id = m.reply_to_message.message_id
    else:
        r_id = m.message_id

    if m.reply_to_message and len(m.text.split()) < 2:
        reason = m.text.split(None, 2)[2]
    elif not m.reply_to_message and len(m.text.split()) < 3:
        reason = m.text.split(None, 2)[2]
    else:
        m.reply_text("Read /help !!")
        return

    if not reason:
        await m.reply_text(
            "You haven't specified a time to mute this user for!",
        )
        return

    split_reason = reason.split(None, 1)
    time_val = split_reason[0].lower()
    if len(split_reason) > 1:
        reason = split_reason[1]
    else:
        reason = ""

    mutetime = await extract_time(m, time_val)

    if not mutetime:
        return

    try:
        await m.chat.restrict_member(
            user_id,
            ChatPermissions(
                can_send_messages=False,
                can_send_media_messages=False,
                can_send_stickers=False,
                can_send_animations=False,
                can_send_games=False,
                can_use_inline_bots=False,
                can_add_web_page_previews=False,
                can_send_polls=False,
                can_change_info=False,
                can_invite_users=True,
                can_pin_messages=False,
            ),
            mutetime,
        )
        LOGGER.info(f"{m.from_user.id} tmuted {user_id} in {m.chat.id}")
        txt = (tlang(m, "admin.mute.muted_user")).format(
            admin=(await mention_html(m.from_user.first_name, m.from_user.id)),
            muted=(await mention_html(user_first_name, user_id)),
        )
        if reason:
            txt += f"\n<b>Reason</b>: {reason}"
        await m.reply_text(txt, reply_to_message_id=r_id)
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except RightForbidden:
        await m.reply_text(tlang(m, "admin.mute.bot_no_right"))
    except UserNotParticipant:
        await m.reply_text(
            "How can I mute a user who is not a part of this chat?",
        )
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)

    return


@Alita.on_message(command("dtmute") & restrict_filter)
async def dtmute_usr(c: Alita, m: Message):
    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text("I can't mute nothing!")
        return

    if not m.reply_to_message:
        return await m.reply_text(
            "No replied message and user to delete and mute!",
        )

    reason = None
    user_id = m.reply_to_message.from_user.id
    user_first_name = m.reply_to_message.from_user.first_name

    if not user_id:
        await m.reply_text("Cannot find user to mute !")
        return
    if user_id == BOT_ID:
        await m.reply_text("Huh, why would I mute myself?")
        return

    if user_id in SUPPORT_STAFF:
        LOGGER.info(
            f"{m.from_user.id} trying to mute {user_id} (SUPPORT_STAFF) in {m.chat.id}",
        )
        await m.reply_text(tlang(m, "admin.support_cannot_restrict"))
        return

    try:
        admins_group = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admins_group = await admin_cache_reload(m, "mute")

    if user_id in admins_group:
        await m.reply_text(tlang(m, "admin.mute.admin_cannot_mute"))
        return

    if m.reply_to_message and len(m.text.split()) < 2:
        reason = m.text.split(None, 2)[2]
    elif not m.reply_to_message and len(m.text.split()) < 3:
        reason = m.text.split(None, 2)[2]
    else:
        m.reply_text("Read /help !!")
        return

    if not reason:
        await m.reply_text(
            "You haven't specified a time to mute this user for!",
        )
        return

    split_reason = reason.split(None, 1)
    time_val = split_reason[0].lower()
    if len(split_reason) > 1:
        reason = split_reason[1]
    else:
        reason = ""

    mutetime = await extract_time(m, time_val)

    if not mutetime:
        return

    try:
        await m.chat.restrict_member(
            user_id,
            ChatPermissions(
                can_send_messages=False,
                can_send_media_messages=False,
                can_send_stickers=False,
                can_send_animations=False,
                can_send_games=False,
                can_use_inline_bots=False,
                can_add_web_page_previews=False,
                can_send_polls=False,
                can_change_info=False,
                can_invite_users=True,
                can_pin_messages=False,
            ),
            mutetime,
        )
        LOGGER.info(f"{m.from_user.id} dtmuted {user_id} in {m.chat.id}")
        await m.reply_to_message.delete()
        txt = (tlang(m, "admin.mute.muted_user")).format(
            admin=(await mention_html(m.from_user.first_name, m.from_user.id)),
            muted=(await mention_html(user_first_name, user_id)),
        )
        if reason:
            txt += f"\n<b>Reason</b>: {reason}"
        await c.send_message(m.chat.id, txt)
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except RightForbidden:
        await m.reply_text(tlang(m, "admin.mute.bot_no_right"))
    except UserNotParticipant:
        await m.reply_text(
            "How can I mute a user who is not a part of this chat?",
        )
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)

    return


@Alita.on_message(command("stmute") & restrict_filter)
async def stmute_usr(c: Alita, m: Message):
    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text("I can't mute nothing!")
        return

    try:
        user_id, user_first_name, _ = await extract_user(c, m)
    except Exception:
        return

    if not user_id:
        await m.reply_text("Cannot find user to mute !")
        return
    if user_id == BOT_ID:
        await m.reply_text("Huh, why would I mute myself?")
        return

    if user_id in SUPPORT_STAFF:
        LOGGER.info(
            f"{m.from_user.id} trying to mute {user_id} (SUPPORT_STAFF) in {m.chat.id}",
        )
        await m.reply_text(tlang(m, "admin.support_cannot_restrict"))
        return

    try:
        admins_group = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admins_group = await admin_cache_reload(m, "mute")

    if user_id in admins_group:
        await m.reply_text(tlang(m, "admin.mute.admin_cannot_mute"))
        return

    if m.reply_to_message and len(m.text.split()) < 2:
        reason = m.text.split(None, 2)[2]
    elif not m.reply_to_message and len(m.text.split()) < 3:
        reason = m.text.split(None, 2)[2]
    else:
        m.reply_text("Read /help !!")
        return

    if not reason:
        await m.reply_text(
            "You haven't specified a time to mute this user for!",
        )
        return

    split_reason = reason.split(None, 1)
    time_val = split_reason[0].lower()
    if len(split_reason) > 1:
        reason = split_reason[1]
    else:
        reason = ""

    mutetime = await extract_time(m, time_val)

    if not mutetime:
        return

    try:
        await m.chat.restrict_member(
            user_id,
            ChatPermissions(
                can_send_messages=False,
                can_send_media_messages=False,
                can_send_stickers=False,
                can_send_animations=False,
                can_send_games=False,
                can_use_inline_bots=False,
                can_add_web_page_previews=False,
                can_send_polls=False,
                can_change_info=False,
                can_invite_users=True,
                can_pin_messages=False,
            ),
            mutetime,
        )
        LOGGER.info(f"{m.from_user.id} stmuted {user_id} in {m.chat.id}")
        await m.delete()
        if m.reply_to_message:
            await m.reply_to_message.delete()
            return
        return
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except RightForbidden:
        await m.reply_text(tlang(m, "admin.mute.bot_no_right"))
    except UserNotParticipant:
        await m.reply_text(
            "How can I mute a user who is not a part of this chat?",
        )
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)

    return


@Alita.on_message(command("mute") & restrict_filter)
async def mute_usr(c: Alita, m: Message):
    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text("I can't mute nothing!")
        return

    reason = None
    if m.reply_to_message:
        r_id = m.reply_to_message.message_id
        if len(m.text.split()) >= 2:
            reason = m.text.split(None, 1)[1]
    else:
        r_id = m.message_id
        if len(m.text.split()) >= 3:
            reason = m.text.split(None, 2)[2]
    try:
        user_id, user_first_name, _ = await extract_user(c, m)
    except Exception:
        return

    if not user_id:
        await m.reply_text("Cannot find user to mute")
        return
    if user_id == BOT_ID:
        await m.reply_text("Huh, why would I mute myself?")
        return

    if user_id in SUPPORT_STAFF:
        LOGGER.info(
            f"{m.from_user.id} trying to mute {user_id} (SUPPORT_STAFF) in {m.chat.id}",
        )
        await m.reply_text(tlang(m, "admin.support_cannot_restrict"))
        return

    try:
        admins_group = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admins_group = await admin_cache_reload(m, "mute")

    if user_id in admins_group:
        await m.reply_text(tlang(m, "admin.mute.admin_cannot_mute"))
        return

    try:
        await m.chat.restrict_member(
            user_id,
            ChatPermissions(
                can_send_messages=False,
                can_send_media_messages=False,
                can_send_stickers=False,
                can_send_animations=False,
                can_send_games=False,
                can_use_inline_bots=False,
                can_add_web_page_previews=False,
                can_send_polls=False,
                can_change_info=False,
                can_invite_users=True,
                can_pin_messages=False,
            ),
        )
        LOGGER.info(f"{m.from_user.id} muted {user_id} in {m.chat.id}")
        txt = (tlang(m, "admin.mute.muted_user")).format(
            admin=(await mention_html(m.from_user.first_name, m.from_user.id)),
            muted=(await mention_html(user_first_name, user_id)),
        )
        if reason:
            txt += f"\n<b>Reason</b>: {reason}"
        await m.reply_text(txt, reply_to_message_id=r_id)
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except RightForbidden:
        await m.reply_text(tlang(m, "admin.mute.bot_no_right"))
    except UserNotParticipant:
        await m.reply_text(
            "How can I mute a user who is not a part of this chat?",
        )
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)

    return


@Alita.on_message(command("smute") & restrict_filter)
async def smute_usr(c: Alita, m: Message):
    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text("I can't mute nothing!")
        return

    reason = None
    if m.reply_to_message:
        r_id = m.reply_to_message.message_id
        if len(m.text.split()) >= 2:
            reason = m.text.split(None, 1)[1]
    else:
        r_id = m.message_id
        if len(m.text.split()) >= 3:
            reason = m.text.split(None, 2)[2]
    try:
        user_id, user_first_name, _ = await extract_user(c, m)
    except Exception:
        return

    if not user_id:
        await m.reply_text("Cannot find user to mute")
        return
    if user_id == BOT_ID:
        await m.reply_text("Huh, why would I mute myself?")
        return

    if user_id in SUPPORT_STAFF:
        LOGGER.info(
            f"{m.from_user.id} trying to mute {user_id} (SUPPORT_STAFF) in {m.chat.id}",
        )
        await m.reply_text(tlang(m, "admin.support_cannot_restrict"))
        return

    try:
        admins_group = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admins_group = await admin_cache_reload(m, "mute")

    if user_id in admins_group:
        await m.reply_text(tlang(m, "admin.mute.admin_cannot_mute"))
        return

    try:
        await m.chat.restrict_member(
            user_id,
            ChatPermissions(
                can_send_messages=False,
                can_send_media_messages=False,
                can_send_stickers=False,
                can_send_animations=False,
                can_send_games=False,
                can_use_inline_bots=False,
                can_add_web_page_previews=False,
                can_send_polls=False,
                can_change_info=False,
                can_invite_users=True,
                can_pin_messages=False,
            ),
        )
        LOGGER.info(f"{m.from_user.id} smuted {user_id} in {m.chat.id}")
        await m.delete()
        if m.reply_to_message:
            await m.reply_to_message.delete()
            return
        return
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except RightForbidden:
        await m.reply_text(tlang(m, "admin.mute.bot_no_right"))
    except UserNotParticipant:
        await m.reply_text(
            "How can I mute a user who is not a part of this chat?",
        )
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)

    return


@Alita.on_message(command("dmute") & restrict_filter)
async def dmute_usr(c: Alita, m: Message):
    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text("I can't mute nothing!")
        return
    if not m.reply_to_message:
        return await m.reply_text(
            "No replied message and user to delete and mute!",
        )

    reason = None
    user_id = m.reply_to_message.from_user.id
    user_first_name = m.reply_to_message.from_user.first_name

    if not user_id:
        await m.reply_text("Cannot find user to mute")
        return
    if user_id == BOT_ID:
        await m.reply_text("Huh, why would I mute myself?")
        return

    if user_id in SUPPORT_STAFF:
        LOGGER.info(
            f"{m.from_user.id} trying to mute {user_id} (SUPPORT_STAFF) in {m.chat.id}",
        )
        await m.reply_text(tlang(m, "admin.support_cannot_restrict"))
        return

    try:
        admins_group = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admins_group = await admin_cache_reload(m, "mute")

    if user_id in admins_group:
        await m.reply_text(tlang(m, "admin.mute.admin_cannot_mute"))
        return

    try:
        await m.chat.restrict_member(
            user_id,
            ChatPermissions(
                can_send_messages=False,
                can_send_media_messages=False,
                can_send_stickers=False,
                can_send_animations=False,
                can_send_games=False,
                can_use_inline_bots=False,
                can_add_web_page_previews=False,
                can_send_polls=False,
                can_change_info=False,
                can_invite_users=True,
                can_pin_messages=False,
            ),
        )
        LOGGER.info(f"{m.from_user.id} dmuted {user_id} in {m.chat.id}")
        await m.reply_to_message.delete()
        txt = (tlang(m, "admin.mute.muted_user")).format(
            admin=(await mention_html(m.from_user.first_name, m.from_user.id)),
            muted=(await mention_html(user_first_name, user_id)),
        )
        if reason:
            txt += f"\n<b>Reason</b>: {reason}"
        await c.send_message(m.chat.id, txt)
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except RightForbidden:
        await m.reply_text(tlang(m, "admin.mute.bot_no_right"))
    except UserNotParticipant:
        await m.reply_text(
            "How can I mute a user who is not a part of this chat?",
        )
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)

    return


@Alita.on_message(command("unmute") & restrict_filter)
async def unmute_usr(c: Alita, m: Message):
    if len(m.text.split()) == 1 and not m.reply_to_message:
        await m.reply_text("I can't unmute nothing!")
        return

    try:
        user_id, user_first_name, _ = await extract_user(c, m)
    except Exception:
        return

    if user_id == BOT_ID:
        await m.reply_text(
            "Huh, why would I unmute myself if you are using me?",
        )
        return

    try:
        await m.chat.restrict_member(user_id, m.chat.permissions)
        LOGGER.info(f"{m.from_user.id} unmuted {user_id} in {m.chat.id}")
        await m.reply_text(
            (tlang(m, "admin.unmute.unmuted_user")).format(
                admin=(await mention_html(m.from_user.first_name, m.from_user.id)),
                unmuted=(await mention_html(user_first_name, user_id)),
            ),
        )
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except UserNotParticipant:
        await m.reply_text(
            "How can I unmute a user who is not a part of this chat?",
        )
    except RightForbidden:
        await m.reply_text(tlang(m, "admin.unmute.bot_no_right"))
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)
    return


__PLUGIN__ = "muting"

__alt_name__ = [
    "mute",
    "tmute",
    "unmute",
]
