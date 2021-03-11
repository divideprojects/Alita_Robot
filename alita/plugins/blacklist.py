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

from alita import PREFIX_HANDLER
from alita.bot_class import Alita
from alita.database.approve_db import Approve
from alita.database.blacklist_db import Blacklist
from alita.tr_engine import tlang
from alita.utils.custom_filters import admin_filter, owner_filter
from alita.utils.parser import mention_html

__PLUGIN__ = "plugins.blacklist.main"
__help__ = "plugins.blacklist.help"


# Initialise
db = Blacklist()
app_db = Approve()


@Alita.on_message(
    filters.command("blacklist", PREFIX_HANDLER) & filters.group & admin_filter,
)
async def view_blacklist(_, m: Message):

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


@Alita.on_message(
    filters.command("addblacklist", PREFIX_HANDLER) & filters.group & admin_filter,
)
async def add_blacklist(_, m: Message):

    if len(m.text.split()) >= 2:
        bl_word = (m.text.split(None, 1)[1]).lower()
        all_blacklisted = db.get_blacklists(m.chat.id)

        if bl_word in all_blacklisted:
            await m.reply_text(
                (tlang(m, "blacklist.already_exists")).format(
                    trigger=bl_word,
                ),
            )
            return

        db.add_blacklist(m.chat.id, bl_word)
        await m.reply_text(
            (tlang(m, "blacklist.added_blacklist")).format(
                trigger=bl_word,
            ),
        )
        return
    await m.reply_text(tlang(m, "general.check_help"))
    return


@Alita.on_message(
    filters.command(["rmblacklist", "unblacklist"], PREFIX_HANDLER)
    & filters.group
    & admin_filter,
)
async def rm_blacklist(_, m: Message):

    if not len(m.text.split()) >= 2:
        await m.reply_text(tlang(m, "general.check_help"))
        return

    chat_bl = db.get_blacklists(m.chat.id)
    bl_word = (m.text.split(None, 1)[1]).lower()
    if bl_word not in chat_bl:
        await m.reply_text(
            (tlang(m, "blacklist.no_bl_found")).format(
                bl_word=f"<code>{bl_word}</code>",
            ),
        )
        return

    db.remove_blacklist(m.chat.id, bl_word)
    await m.reply_text(
        (tlang(m, "blacklist.rm_blacklist")).format(
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
            (tlang(m, "blacklist.action_set")).format(action=action),
        )
    elif len(m.text.split()) == 1:
        action = db.get_action(m.chat.id)
        await m.reply_text(
            (tlang(m, "blacklist.action_get")).format(action=action),
        )
    else:
        await m.reply_text(tlang(m, "general.check_help"))

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
    await q.answer("Cleared all Blacklists!", show_alert=True)
    return
