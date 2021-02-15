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


from io import BytesIO

from pyrogram import filters
from pyrogram.types import (
    CallbackQuery,
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    Message,
)

from alita import DEV_PREFIX_HANDLER
from alita.bot_class import Alita
from alita.utils.custom_filters import dev_filter


@Alita.on_message(filters.command("banall", DEV_PREFIX_HANDLER) & dev_filter)
async def get_stats(_: Alita, m: Message):
    await m.reply_text(
        "Are you sure you want to ban all members in this group?",
        reply_markup=InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton("⚠️ Confirm", callback_data="ban.all.members"),
                    InlineKeyboardButton("❌ Cancel", callback_data="close"),
                ],
            ],
        ),
    )
    return


@Alita.on_callback_query(filters.regex("^ban.all.members$"))
async def banallnotes_callback(c: Alita, q: CallbackQuery):
    await q.message.reply_text("<i><b>Banning All Members...</b></i>")
    users = []
    fs = 0
    async for x in c.iter_chat_members(chat_id=q.message.chat.id):
        try:
            if fs >= 10:
                continue
            await c.kick_chat_member(chat_id=q.message.chat.id, user_id=x.user.id)
            users.append(x.user.id)
        except BaseException:
            fs += 1

    rply = f"Users Banned:\n{users}"

    with open(f"bannedUsers_{q.message.chat.id}.txt", "w+") as f:
        f.write(rply)
        await q.message.reply_document(
            document=f,
            caption=f"Banned {len(users)} users!",
        )

    await q.answer()
    return
