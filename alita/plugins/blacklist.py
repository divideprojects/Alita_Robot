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


from html import escape
from time import time

from pyrogram import filters
from pyrogram.errors import RPCError
from pyrogram.types import (
    CallbackQuery,
    ChatPermissions,
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    Message,
)

from alita import LOGGER, PREFIX_HANDLER
from alita.bot_class import Alita
from alita.database.approve_db import Approve
from alita.database.blacklist_db import Blacklist
from alita.tr_engine import tlang
from alita.utils.custom_filters import admin_filter, owner_filter
from alita.utils.parser import mention_html
from alita.utils.regex_utils import regex_searcher

__PLUGIN__ = "Blacklist"

__help__ = """
Want to restrict certain words or sentences in your group?

Blacklists are used to stop certain triggers from being said in a group. Any time the trigger is mentioned, \
the message will immediately be deleted. A good combo is sometimes to pair this up with warn filters!
**NOTE:** blacklists do not affect group admins.
 × /blacklist: View the current blacklisted words.

**Admin only:**
 × /addblacklist <triggers>: Add a trigger to the blacklist. Each line is considered one trigger, so using different \
lines will allow you to add muser_listtiple triggers.
 × /unblacklist <triggers>: Remove triggers from the blacklist. Same newline logic applies here, so you can remove \
muser_listtiple triggers at once.
 × /rmblacklist <triggers>: Same as above.
 × /blaction <action>: This action will occur when user uses a blacklist word. Choose from - 'kick', 'ban', 'mute', 'warn'
 Default is 'kick', which will kick the user on typing blacklist word.

**Owner Only**
 × /rmallblacklist: Removes all the blacklists from the current chat.

**Note:** Can only add or remove one blacklist at a time!
"""

# Initialise
db = Blacklist()
app_db = Approve()


@Alita.on_message(
    filters.command("blacklist", PREFIX_HANDLER) & filters.group & admin_filter,
)
async def view_blacklist(_, m: Message):

    chat_title = m.chat.title
    blacklists_chat = (await tlang(m, "blacklist.curr_blacklist_initial")).format(
        chat_title=f"<b>{chat_title}</b>",
    )
    all_blacklisted = db.get_blacklists(m.chat.id)

    if not all_blacklisted:
        await m.reply_text(
            (await tlang(m, "blacklist.no_blacklist")).format(
                chat_title=f"<b>{chat_title}</b>",
            ),
        )
        return

    blacklists_chat += "\n".join(
        [f" • <code>{escape(i)}</code>" for i in all_blacklisted],
    )

    await m.reply_text(blacklists_chat)
    return


@Alita.on_message(
    filters.command("addblacklist", PREFIX_HANDLER) & filters.group & admin_filter,
)
async def add_blacklist(_, m: Message):

    if len(m.text.split()) >= 2:
        bl_word = (m.text.split(None, 1)[1]).lower()
        db.add_blacklist(m.chat.id, bl_word)
        await m.reply_text(
            (await tlang(m, "blacklist.added_blacklist")).format(
                trigger=f"<code>{bl_word}</code>",
            ),
        )
        return
    await m.reply_text(await tlang(m, "general.check_help"))
    return


@Alita.on_message(
    filters.command(["rmblacklist", "unblacklist"], PREFIX_HANDLER)
    & filters.group
    & admin_filter,
)
async def rm_blacklist(_, m: Message):

    chat_bl = db.get_blacklists(m.chat.id)
    if not len(m.text.split()) >= 2:
        await m.reply_text(await tlang(m, "general.check_help"))
        return

    bl_word = (m.text.split(None, 1)[1]).lower()
    if bl_word not in chat_bl:
        await m.reply_text(
            (await tlang(m, "blacklist.no_bl_found")).format(
                bl_word=f"<code>{bl_word}</code>",
            ),
        )
        return

    db.remove_blacklist(m.chat.id, bl_word)
    await m.reply_text(
        (await tlang(m, "blacklist.rm_blacklist")).format(
            bl_word=f"<code>{bl_word}</code>",
        ),
    )

    return


@Alita.on_message(filters.command("blaction", PREFIX_HANDLER) & filters.group)
async def set_bl_action(_, m: Message):
    if len(m.text.split()) == 2:
        action = m.text.split(None, 1)[1]
        db.set_action(m.chat.id, action)
        await m.reply_text(
            (await tlang(m, "blacklist.action_set")).format(action=f"<b>{action}</b>"),
        )
    elif len(m.text.split()) == 1:
        action = db.get_action(m.chat.id)
        await m.reply_text(
            (await tlang(m, "blacklist.action_get")).format(action=f"<b>{action}</b>"),
        )
    else:
        await m.reply_text(await tlang(m, "general.check_help"))

    return


@Alita.on_message(filters.group, group=3)
async def del_blacklist(_, m: Message):

    if not m.from_user:
        return

    chat_blacklists = db.get_blacklists(m.chat.id)
    action = db.get_action(m.chat.id)

    # If no blacklists, then return
    if not chat_blacklists:
        return

    try:
        approved_users = []
        app_users = app_db.list_approved(m.chat.id)

        for i in app_users:
            approved_users.append(int(i["user_id"]))

        async for i in m.chat.iter_members(filter="administrators"):
            approved_users.append(i.user.id)

        # If user_id in approved_users list, return and don't delete the message
        if m.from_user.id in approved_users:
            return

        if m.text:
            for trigger in chat_blacklists:
                pattern = r"( |^|[^\w])" + trigger + r"( |$|[^\w])"
                match = await regex_searcher(pattern, m.text.lower())
                if not match:
                    continue
                if match:
                    try:
                        await perform_action(m, action)
                        await m.delete()
                    except RPCError as ef:
                        LOGGER.info(ef)
                    break

    except AttributeError:
        pass  # Skip attribute errors!


# TODO - Add warn option when Warn db is added!!
async def perform_action(m: Message, action: str):
    if action == "kick":
        (await m.chat.kick_member(m.from_user.id, int(time() + 45)))
        await m.reply_text(
            (
                f"Kicked {m.from_user.username if m.from_user.username else m.from_user.first_name}"
                " for using a blacklisted word!"
            ),
        )
    elif action == "ban":
        (
            await m.chat.kick_member(
                m.from_user.id,
            )
        )
        await m.reply_text(
            (
                f"Banned {m.from_user.username if m.from_user.username else m.from_user.first_name}"
                " for using a blacklisted word!"
            ),
        )
    elif action == "mute":
        (
            await m.chat.restrict_member(
                m.from_user.id,
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
        )
        await m.reply_text(
            (
                f"Muted {m.from_user.username if m.from_user.username else m.from_user.first_name}"
                " for using a blacklisted word!"
            ),
        )
    return


@Alita.on_message(
    filters.command("rmallblacklist", PREFIX_HANDLER) & filters.group & owner_filter,
)
async def rm_allblacklist(_, m: Message):

    all_bls = db.get_blacklists(m.chat.id)
    if not all_bls:
        await m.reply_text("No notes are blacklists in this chat")
        return

    await m.reply_text(
        "Are you sure you want to clear all blacklists?",
        reply_markup=InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton(
                        "⚠️ Confirm",
                        callback_data=f"rm_allbl.{m.from_user.id}.{m.from_user.first_name}",
                    ),
                    InlineKeyboardButton("❌ Cancel", callback_data="close"),
                ],
            ],
        ),
    )
    return


@Alita.on_callback_query(filters.regex("^rm_allbl."))
async def rm_allbl_callback(_, q: CallbackQuery):
    user_id = q.data.split(".")[-2]
    name = q.data.split(".")[-1]
    user_status = (await q.message.chat.get_member(user_id)).status
    if user_status != "creator":
        await q.message.edit(
            (
                f"You're an admin {await mention_html(name, user_id)}, not owner!\n"
                "Stay in your limits!"
            ),
        )
        return
    db.rm_all_blacklist(q.message.chat.id)
    await q.message.delete()
    await q.answer("Cleared all notes!", show_alert=True)
    return
