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
from pyrogram.errors import PeerIdInvalid, RPCError, UserNotParticipant
from pyrogram.types import (
    CallbackQuery,
    ChatPermissions,
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    Message,
)

from alita import LOGGER, PREFIX_HANDLER, SUPPORT_GROUP
from alita.bot_class import Alita
from alita.database.approve_db import Approve
from alita.utils.custom_filters import admin_filter, owner_filter
from alita.utils.extract_user import extract_user
from alita.utils.parser import mention_html

# Initialise
db = Approve()


@Alita.on_message(
    filters.command("approve", PREFIX_HANDLER) & admin_filter,
)
async def approve_user(c: Alita, m: Message):

    chat_title = m.chat.title
    chat_id = m.chat.id
    user_id, user_first_name, _ = await extract_user(c, m)
    if not user_id:
        await m.reply_text(
            "I don't know who you're talking about, you're going to need to specify a user!",
        )
        return
    try:
        member = await m.chat.get_member(user_id)
    except UserNotParticipant:
        await m.reply_text("This user is not in this chat!")
        return

    except RPCError as ef:
        await m.reply_text(
            f"<b>Error</b>: <code>{ef}</code>\nReport it to @{SUPPORT_GROUP}",
        )
        return
    if member.status in ("administrator", "creator"):
        await m.reply_text(
            "User is already admin - blacklists and locks already don't apply to them.",
        )
        return
    already_approved = db.check_approve(chat_id, user_id)
    if already_approved:
        await m.reply_text(
            f"{(await mention_html(user_first_name, user_id))} is already approved in {chat_title}",
        )
        return
    db.add_approve(chat_id, user_id, user_first_name)
    LOGGER.info(f"{user_id} approved by {m.from_user.id} in {m.chat.id}")

    # Allow all permissions
    await m.chat.restrict_member(
        user_id=user_id,
        permissions=ChatPermissions(
            can_send_messages=True,
            can_send_media_messages=True,
            can_send_stickers=True,
            can_send_animations=True,
            can_send_games=True,
            can_use_inline_bots=True,
            can_add_web_page_previews=True,
            can_send_polls=True,
            can_change_info=True,
            can_invite_users=True,
            can_pin_messages=True,
        ),
    )

    await m.reply_text(
        (
            f"{(await mention_html(user_first_name, user_id))} has been approved in {chat_title}!\n"
            "They will now be ignored by blacklists, locks and antiflood!"
        ),
    )
    return


@Alita.on_message(
    filters.command(["disapprove", "unapprove"], PREFIX_HANDLER) & admin_filter,
)
async def disapprove_user(c: Alita, m: Message):

    chat_title = m.chat.title
    chat_id = m.chat.id
    user_id, user_first_name, _ = await extract_user(c, m)
    already_approved = db.check_approve(chat_id, user_id)
    if not user_id:
        await m.reply_text(
            "I don't know who you're talking about, you're going to need to specify a user!",
        )
        return
    try:
        member = await m.chat.get_member(user_id)
    except UserNotParticipant:
        if already_approved:  # If user is approved and not in chat, unapprove them.
            db.remove_approve(chat_id, user_id)
            LOGGER.info(f"{user_id} disapproved in {m.chat.id} as UserNotParticipant")
        await m.reply_text("This user is not in this chat, unapproved them.")
        return
    except RPCError as ef:
        await m.reply_text(
            f"<b>Error</b>: <code>{ef}</code>\nReport it to @{SUPPORT_GROUP}",
        )
        return

    if member.status in ("administrator", "creator"):
        await m.reply_text("This user is an admin, they can't be disapproved.")
        return

    if not already_approved:
        await m.reply_text(
            f"{(await mention_html(user_first_name, user_id))} isn't approved yet!",
        )
        return

    db.remove_approve(chat_id, user_id)
    LOGGER.info(f"{user_id} disapproved by {m.from_user.id} in {m.chat.id}")

    # Set permission same as of current user by fetching them from chat!
    await m.chat.restrict_member(
        user_id=user_id,
        permissions=m.chat.permissions,
    )

    await m.reply_text(
        f"{(await mention_html(user_first_name, user_id))} is no longer approved in {chat_title}.",
    )
    return


@Alita.on_message(filters.command("approved", PREFIX_HANDLER) & admin_filter)
async def check_approved(_, m: Message):

    chat = m.chat
    chat_title = chat.title
    msg = "The following users are approved:\n"
    approved_people = db.list_approved(chat.id)

    if not approved_people:
        await m.reply_text(f"No users are approved in {chat_title}.")
        return

    for user_id, user_name in approved_people:
        try:
            await chat.get_member(user_id)  # Check if user is in chat or not
        except UserNotParticipant:
            db.remove_approve(chat.id, user_id)
            continue
        except PeerIdInvalid:
            pass
        msg += f"- `{user_id}`: {user_name}\n"
    await m.reply_text(msg)
    LOGGER.info(f"{m.from_user.id} checking approved users in {m.chat.id}")
    return


@Alita.on_message(filters.command("approval", PREFIX_HANDLER) & admin_filter)
async def check_approval(c: Alita, m: Message):

    user_id, user_first_name, _ = await extract_user(c, m)
    check_approve = db.check_approve(m.chat.id, user_id)
    LOGGER.info(f"{m.from_user.id} checking approval of {user_id} in {m.chat.id}")

    if not user_id:
        await m.reply_text(
            "I don't know who you're talking about, you're going to need to specify a user!",
        )
        return
    if check_approve:
        await m.reply_text(
            f"{(await mention_html(user_first_name, user_id))} is an approved user. Locks, antiflood, and blacklists won't apply to them.",
        )
    else:
        await m.reply_text(
            f"{(await mention_html(user_first_name, user_id))} is not an approved user. They are affected by normal commands.",
        )
    return


@Alita.on_message(
    filters.command("unapproveall", PREFIX_HANDLER) & filters.group & owner_filter,
)
async def unapproveall_users(_, m: Message):

    all_approved = db.list_approved(m.chat.id)
    if not all_approved:
        await m.reply_text("No one is approved in this chat.")
        return

    await m.reply_text(
        "Are you sure you want to remove everyone who is approved in this chat?",
        reply_markup=InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton(
                        "⚠️ Confirm",
                        callback_data=f"unapprove_all.{m.from_user.id}.{m.from_user.first_name}",
                    ),
                    InlineKeyboardButton("❌ Cancel", callback_data="close"),
                ],
            ],
        ),
    )
    return


@Alita.on_callback_query(filters.regex("^unapprove_all."))
async def unapproveall_callback(_, q: CallbackQuery):
    user_id = q.data.split(".")[-2]
    name = q.data.split(".")[-1]
    approved_people = db.list_approved(q.message.chat.id)
    user_status = (await q.message.chat.get_member(user_id)).status
    if user_status != "creator":
        await q.message.edit(
            (
                f"You're an admin {await mention_html(name, user_id)}, not owner!\n"
                "Stay in your limits!"
            ),
        )
        return
    db.unapprove_all(q.message.chat.id)
    for i in approved_people:
        await q.message.chat.restrict_member(
            user_id=i,
            permissions=q.message.chat.permissions,
        )
    await q.message.delete()
    LOGGER.info(f"{user_id} disapproved all users in {q.message.chat.id}")
    await q.answer("Disapproved all users!", show_alert=True)
    return


__PLUGIN__ = "plugins.approve.main"
__help__ = "plugins.approve.help"
__alt_name__ = ["approved"]
