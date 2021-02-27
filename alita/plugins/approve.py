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
from pyrogram.errors import RPCError, UserNotParticipant
from pyrogram.types import (
    CallbackQuery,
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    Message,
)

from alita import PREFIX_HANDLER, SUPPORT_GROUP
from alita.bot_class import Alita
from alita.database.approve_db import Approve
from alita.utils.custom_filters import admin_filter
from alita.utils.extract_user import extract_user
from alita.utils.parser import mention_html

# Initialise
db = Approve()

__PLUGIN__ = "Approve"

__help__ = """
Sometimes, you might trust a user not to send unwanted content.
Maybe not enough to make them admin, but you might be ok with locks, blacklists, and antiflood not applying to them.
That's what approvals are for - approve trustworthy users to allow them to send stuff without restrictions!

**Admin commands:**
 × /approval: Check a user's approval status in this chat.
 × /approve: Approve of a user. Locks, blacklists, and antiflood won't apply to them anymore.
 × /unapprove: Unapprove of a user. They will now be subject to blocklists.
 × /approved: List all approved users.
 × /unapproveall: Unapprove *ALL* users in a chat. This cannot be undone!
"""


@Alita.on_message(
    filters.command("approve", PREFIX_HANDLER) & filters.group & admin_filter,
)
async def approve_user(_, m: Message):

    chat_title = m.chat.title
    chat_id = m.chat.id
    user_id, user_first_name = await extract_user(m)
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
    if await db.check_approve(chat_id, user_id):
        await m.reply_text(
            f"{(await mention_html(user_first_name, user_id))} is already approved in {chat_title}",
        )
        return
    await db.add_approve(chat_id, user_id)
    await m.reply_text(
        f"{(await mention_html(user_first_name, user_id))} has been approved in {chat_title}! They will now be ignored by blocklists.",
    )
    return


@Alita.on_message(
    filters.command("disapprove", PREFIX_HANDLER) & filters.group & admin_filter,
)
async def disapprove_user(_, m: Message):

    chat_title = m.chat.title
    chat_id = m.chat.id

    user_id, user_first_name = await extract_user(m)
    if not user_id:
        await m.reply_text(
            "I don't know who you're talking about, you're going to need to specify a user!",
        )
        return
    try:
        member = await m.chat.get_member(user_id)
    except UserNotParticipant:
        if await db.check_approve(
            chat_id,
            user_id,
        ):  # If user is approved and not in chat, unapprove them.
            await db.remove_approve(chat_id, user_id)
        await m.reply_text("This user is not in this chat, unapproved them.")
        return
    except RPCError as ef:
        await m.reply_text(
            f"<b>Error</b>: <code>{ef}</code>\nReport it to @{SUPPORT_GROUP}",
        )
        return
    if member.status in ["administrator", "creator"]:
        await m.reply_text("This user is an admin, they can't be unapproved.")
        return
    if not (await db.check_approve(chat_id, user_id)):
        await m.reply_text(
            f"{(await mention_html(user_first_name, user_id))} isn't approved yet!",
        )
        return
    await db.remove_approve(chat_id, user_id)
    await m.reply_text(
        f"{(await mention_html(user_first_name, user_id))} is no longer approved in {chat_title}.",
    )
    return


@Alita.on_message(
    filters.command("approved", PREFIX_HANDLER) & filters.group & admin_filter,
)
async def check_approved(_, m: Message):

    chat = m.chat
    chat_title = chat.title
    user_id = (await extract_user(m))[0]
    msg = "The following users are approved:\n"
    approved_people = await db.list_approved(chat.id)

    if not approved_people:
        await m.reply_text(f"No users are approved in {chat_title}.")
        return

    for i in approved_people:
        try:
            member = await chat.get_member(int(i.user_id))
        except UserNotParticipant:
            await db.remove_approve(chat.id, user_id)
            continue
        msg += f"- `{i.user_id}`: {(await mention_html(member.user['first_name'], int(i.user_id)))}\n"
    await m.reply_text(msg)
    return


@Alita.on_message(
    filters.command("approval", PREFIX_HANDLER) & filters.group & admin_filter,
)
async def check_approval(_, m: Message):

    user_id, user_first_name = await extract_user(m)
    check_approve = await db.check_approve(m.chat.id, user_id)

    if not user_id:
        await m.reply_text(
            "I don't know who you're talking about, you're going to need to specify a user!",
        )
        return
    if check_approve:
        await m.reply_text(
            f"{(await mention_html(user_first_name, user_id))} is an approved user. Locks, antiflood, and blocklists won't apply to them.",
        )
    else:
        await m.reply_text(
            f"{(await mention_html(user_first_name, user_id))} is not an approved user. They are affected by normal commands.",
        )
    return


@Alita.on_message(
    filters.command("unapproveall", PREFIX_HANDLER) & filters.group,
)
async def unapproveall_users(_, m: Message):

    all_approved = await db.list_approved(m.chat.id)
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
                        callback_data=f"unapprove.all.{m.from_user.id}.{m.from_user.first_name}",
                    ),
                    InlineKeyboardButton("❌ Cancel", callback_data="close"),
                ],
            ],
        ),
    )
    return


@Alita.on_message(filters.regex("^unapprove.all.") & filters.group)
async def unapproveall_callback(_, q: CallbackQuery):
    user_id = q.data.split(".")[-2]
    name = q.data.split(".")[-1]
    if (await q.message.chat.get_member(user_id)).status != "creator":
        await q.message.edit(
            (
                f"You're an admin {await mention_html(name, user_id)}, not owner!\n"
                "Stay in your limits!"
            ),
        )
        return

    try:
        await db.unapprove_all(q.message.chat.id)
        await q.message.edit_text(
            f"All users have been disapproved in {q.message.chat.title}",
        )
    except RPCError as ef:
        await q.message.edit_text(
            f"Some Error occured, report at @{SUPPORT_GROUP}.\n{ef}",
        )
    return
