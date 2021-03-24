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
from pyrogram.types import (
    CallbackQuery,
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    Message,
)

from alita import LOGGER, PREFIX_HANDLER
from alita.bot_class import Alita
from alita.tr_engine import tlang


async def gen_formatting_kb(m):
    keyboard = InlineKeyboardMarkup(
        [
            [
                InlineKeyboardButton(
                    "Markdown Formatting",
                    callback_data="formatting.md_formatting",
                ),
                InlineKeyboardButton("Fillings", callback_data="formatting.fillings"),
            ],
            [
                InlineKeyboardButton(
                    "Random Content",
                    callback_data="formatting.random_content",
                ),
            ],
            [
                InlineKeyboardButton(
                    ("« " + (tlang(m, "general.back_btn"))),
                    callback_data="commands",
                ),
            ],
        ],
    )
    return keyboard


@Alita.on_message(
    filters.command(["markdownhelp", "formatting"], PREFIX_HANDLER) & filters.private,
)
async def markdownhelp(_, m: Message):
    await m.reply_text(
        tlang(m, __help__),
        quote=True,
        reply_markup=(await gen_formatting_kb(m)),
    )
    LOGGER.info(f"{m.from_user.id} used cmd '{m.command[0]}' in {m.chat.id}")
    return


@Alita.on_callback_query(filters.regex("^formatting."))
async def get_formatting_info(_, q: CallbackQuery):
    cmd = q.data.split(".")[1]
    kb = InlineKeyboardMarkup(
        [
            [
                InlineKeyboardButton(
                    (tlang(q, "general.back_btn")),
                    callback_data="back.formatting",
                ),
            ],
        ],
    )

    if cmd == "md_formatting":
        await q.message.edit_text(
            tlang(q, "formatting.md_help"),
            reply_markup=kb,
            parse_mode="html",
        )
    elif cmd == "fillings":
        await q.message.edit_text(
            tlang(q, "formatting.filling_help"),
            reply_markup=kb,
            parse_mode="html",
        )
    elif cmd == "random_content":
        await q.message.edit_text(
            tlang(q, "formatting.random_help"),
            reply_markup=kb,
            parse_mode="html",
        )

    await q.answer()
    return


@Alita.on_callback_query(filters.regex("^back."))
async def send_mod_help(_, q: CallbackQuery):
    await q.message.edit_text(
        (tlang(q, "plugins.formatting.help")),
        reply_markup=(await gen_formatting_kb(q.message)),
    )
    await q.answer()
    return


__PLUGIN__ = "plugins.formatting.main"
__help__ = "plugins.formatting.help"
__alt_name__ = ["formatting", "markdownhelp", "markdown"]
__buttons__ = [
    [
        InlineKeyboardButton(
            "Markdown Formatting",
            callback_data="formatting.md_formatting",
        ),
        InlineKeyboardButton("Fillings", callback_data="formatting.fillings"),
    ],
    [
        InlineKeyboardButton(
            "Random Content",
            callback_data="formatting.random_content",
        ),
    ],
]
