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

from pyrogram import filters
from pyrogram.types import (
    CallbackQuery,
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    Message,
)

from alita import LOGGER, PREFIX_HANDLER
from alita.bot_class import Alita
from alita.database.approve_db import Approve
from alita.database.blacklist_db import Blacklist
from alita.tr_engine import tlang
from alita.utils.custom_filters import owner_filter, restrict_filter
from alita.utils.parser import mention_html

# Initialise
db = Blacklist()
app_db = Approve()


@Alita.on_message(filters.command("blacklist", PREFIX_HANDLER) & filters.group)
async def view_blacklist(_, m: Message):

    LOGGER.info(f"{m.from_user.id} checking blacklists in {m.chat.id}")

    chat_title = m.chat.title
    blacklists_chat = (tlang(m, "blacklist.curr_blacklist_initial")).format(
        chat_title=chat_title,
    )
    all_blacklisted = db.get_blacklists(m.chat.id)

    if not all_blacklisted:
        await m.reply_text(
            (tlang(m, "blacklist.no_blacklist")).format(
                chat_title=chat_title,
            ),
        )
        return

    blacklists_chat += "\n".join(
        [f" • <code>{escape(i)}</code>" for i in all_blacklisted],
    )

    await m.reply_text(blacklists_chat)
    return


@Alita.on_message(filters.command("addblacklist", PREFIX_HANDLER) & restrict_filter)
async def add_blacklist(_, m: Message):

    if not len(m.text.split()) >= 2:
        await m.reply_text(tlang(m, "general.check_help"))
        return

    bl_words = ((m.text.split(None, 1)[1]).lower()).split()
    all_blacklisted = db.get_blacklists(m.chat.id)
    already_added_words, rep_text = [], ""

    for bl_word in bl_words:
        if bl_word in all_blacklisted:
            already_added_words.append(bl_word)
            continue
        db.add_blacklist(m.chat.id, bl_word)

    if already_added_words:
        rep_text = (
            ", ".join([f"<code>{i}</code>" for i in bl_words])
            + " already added in blacklist, skipped them!"
        )
    LOGGER.info(f"{m.from_user.id} added new blacklists ({bl_words}) in {m.chat.id}")
    await m.reply_text(
        (
            (tlang(m, "blacklist.added_blacklist")).format(
                trigger=", ".join([f"<code>{i}</code>" for i in bl_words]),
            )
            + (f"\n{rep_text}" if rep_text else "")
        ),
    )
    await m.stop_propagation()


@Alita.on_message(
    filters.command(["blwarning", "blreason", "blacklistreason"], PREFIX_HANDLER)
    & restrict_filter,
)
async def blacklistreason(_, m: Message):
    if len(m.text.split()) == 1:
        curr = db.get_reason(m.chat.id)
        await m.reply_text(
            f"The current reason for blacklists warn is:\n<code>{curr}</code>",
        )
    else:
        reason = m.text.split(None, 1)[1]
        db.set_reason(m.chat.id, reason)
        await m.reply_text(
            f"Updated reason for blacklists warn is:\n<code>{reason}</code>",
        )
    return


@Alita.on_message(
    filters.command(["rmblacklist", "unblacklist"], PREFIX_HANDLER) & restrict_filter,
)
async def rm_blacklist(_, m: Message):

    if not len(m.text.split()) >= 2:
        await m.reply_text(tlang(m, "general.check_help"))
        return

    chat_bl = db.get_blacklists(m.chat.id)
    non_found_words, rep_text = [], ""
    bl_words = ((m.text.split(None, 1)[1]).lower()).split()

    for bl_word in bl_words:
        if bl_word not in chat_bl:
            non_found_words.append(bl_word)
            continue
        db.remove_blacklist(m.chat.id, bl_word)

    if non_found_words == bl_words:
        return await m.reply_text("Blacklists not found!")

    if non_found_words:
        rep_text = (
            "Could not find "
            + ", ".join([f"<code>{i}</code>" for i in non_found_words])
            + " in blcklisted words, skipped them."
        )

    LOGGER.info(f"{m.from_user.id} removed blacklists ({bl_words}) in {m.chat.id}")
    await m.reply_text(
        (
            (tlang(m, "blacklist.rm_blacklist")).format(
                bl_words=", ".join([f"<code>{i}</code>" for i in bl_words]),
            )
            + (f"\n{rep_text}" if rep_text else "")
        ),
    )
    await m.stop_propagation()


@Alita.on_message(
    filters.command(["blaction", "blacklistaction", "blacklistmode"], PREFIX_HANDLER)
    & restrict_filter,
)
async def set_bl_action(_, m: Message):
    if len(m.text.split()) == 2:
        action = m.text.split(None, 1)[1]
        valid_actions = ("ban", "kick", "mute", "warn", "none")
        if action not in valid_actions:
            await m.reply_text(
                "Choose a valid blacklist action from "
                + ", ".join([f"<code>{i}</code>" for i in valid_actions]),
            )
            return
        db.set_action(m.chat.id, action)
        LOGGER.info(
            f"{m.from_user.id} set blacklist action to '{action}' in {m.chat.id}",
        )
        await m.reply_text(
            (tlang(m, "blacklist.action_set")).format(action=action),
        )
    elif len(m.text.split()) == 1:
        action = db.get_action(m.chat.id)
        LOGGER.info(f"{m.from_user.id} checking blacklist action in {m.chat.id}")
        await m.reply_text(
            (tlang(m, "blacklist.action_get")).format(action=action),
        )
    else:
        await m.reply_text(tlang(m, "general.check_help"))

    return


@Alita.on_message(
    filters.command("rmallblacklist", PREFIX_HANDLER) & owner_filter,
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
    LOGGER.info(f"{user_id} removed all blacklists in {q.message.chat.id}")
    await q.answer("Cleared all Blacklists!", show_alert=True)
    return


__PLUGIN__ = "plugins.blacklist.main"
__help__ = "plugins.blacklist.help"
__alt_name__ = ["blacklists", "blaction"]
