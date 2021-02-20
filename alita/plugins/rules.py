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
from pyrogram.errors import UserIsBlocked
from pyrogram.types import (
    CallbackQuery,
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    Message,
)

from alita import PREFIX_HANDLER
from alita.bot_class import Alita
from alita.db import rules_db as db
from alita.utils.admin_filter import admin_filter
from alita.utils.localization import GetLang

__PLUGIN__ = "Rules"

__help__ = """
Set rules for you chat so that members know what to do and \
what not to do in your group!

 × /rules: get the rules for current chat.

**Admin only:**
 × /setrules <rules>: Set the rules for this chat, also works as a reply to a message.
 × /clearrules: Clear the rules for this chat.
"""


@Alita.on_message(filters.command("rules", PREFIX_HANDLER) & filters.group)
async def get_rules(c: Alita, m: Message):
    _ = GetLang(m).strs

    chat_id = m.chat.id
    rules = db.get_rules(chat_id)

    if not rules:
        await m.reply_text(_("rules.no_rules"), reply_to_message_id=m.message_id)
        return

    try:
        await c.send_message(
            m.from_user.id,
            _("rules.get_rules").format(chat=m.chat.title, rules=rules),
        )
    except UserIsBlocked:
        me = await c.get_me()
        pm_kb = InlineKeyboardMarkup(
            [[InlineKeyboardButton("PM", url=f"https://t.me/{me.username}?start")]],
        )
        await m.reply_text(
            _("rules.pm_me"),
            reply_to_message_id=m.message_id,
            reply_markup=pm_kb,
        )
        return

    await m.reply_text(_("rules.sent_pm_rules"), reply_to_message_id=m.message_id)
    return


@Alita.on_message(
    filters.command("setrules", PREFIX_HANDLER) & filters.group & admin_filter,
)
async def set_rules(_: Alita, m: Message):

    _ = GetLang(m).strs

    chat_id = m.chat.id
    if m.reply_to_message and m.reply_to_message.text:
        rules = m.reply_to_message.text
    elif (not m.reply_to_message) and len(m.text.split()) >= 2:
        rules = m.text.split(None, 1)[1]

    if len(rules) > 4000:
        rules = rules[0:3949]  # Split Rules if len > 4000 chars
        await m.reply_text("Rules truncated to 3950 characters!")

    db.set_rules(chat_id, rules)
    await m.reply_text(_("rules.set_rules"), reply_to_message_id=m.message_id)
    return


@Alita.on_message(
    filters.command("clearrules", PREFIX_HANDLER) & filters.group & admin_filter,
)
async def clear_rules(_: Alita, m: Message):

    _ = GetLang(m).strs

    rules = db.get_rules(m.chat.id)
    if not rules:
        await m.reply_text(_("rules.no_rules"), reply_to_message_id=m.message_id)
        return

    await m.reply_text(
        "Are you sure you want to clear rules?",
        reply_markup=InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton("⚠️ Confirm", callback_data="clear.rules"),
                    InlineKeyboardButton("❌ Cancel", callback_data="close"),
                ],
            ],
        ),
    )
    return


@Alita.on_callback_query(filters.regex("^clear.rules$"))
async def clearrules_callback(_: Alita, q: CallbackQuery):
    _ = GetLang(q.message).strs
    db.clear_rules(q.message.chat.id)
    await q.message.reply_text(_("rules.clear"))
    await q.answer()
    return
