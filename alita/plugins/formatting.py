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

md_formatting_str = """
**Markdown Formatting**

You can format your message using **bold**, __italics__, --underline--, and much more. Go ahead and experiment!

**Supported markdown**:
- ``code words``: Backticks are used for monospace fonts. Shows as: `code words`.
- `__italic words__`: Underscores are used for italic fonts. Shows as: __italic words__.
- `*bold words*`: Asterisks are used for bold fonts. Shows as: **bold words**.
- `~strikethrough~`: Tildes are used for strikethrough. Shows as: ~~strikethrough~~.
- `[hyperlink](example.com)`: This is the formatting used for hyperlinks. Shows as: [hyperlink](https://example.com/).
- `[My button](buttonurl://example.com)`: This is the formatting used for creating buttons. This example will create a button named "My button" which opens `example.com` when clicked.
If you would like to send buttons on the same row, use the `:same` formatting. EG:
`[button 1](buttonurl://example.com)`
`[button 2](buttonurl://example.com:same)`
`[button 3](buttonurl://example.com)`
This will show button 1 and 2 on the same line, with 3 underneath.
"""

filling_str = """
**Fillings**

You can also customise the contents of your message with contextual data. For example, you could mention a user by name in the welcome message, or mention them in a filter!

Supported fillings:
- {first}: The user's first name.
- {last}: The user's last name.
- {fullname}: The user's full name.
- {username}: The user's username. If they don't have one, mentions the user instead.
- {mention}: Mentions the user with their firstname.
- {id}: The user's ID.
- {chatname}: The chat's name.
"""

random_content_str = """
Random Content

Another thing that can be fun, is to randomise the contents of a message. Make things a little more personal by changing welcome messages, or changing notes!

How to use random contents:
- %%%: This separator can be used to add "random" replies to the bot.
For example:
hello
%%%
how are you
This will randomly choose between sending the first message, "hello", or the second message, "how are you". Use this to make Rose feel a bit more customised! (only works in notes/filters/greetings)

Example welcome message::
- Every time a new user joins, they'll be presented with one of the three messages shown here.
-> /filter "hey"
hello there {first}!
%%%
Ooooh, {first} how are you?
%%%
Sup? {first}"""


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
    filters.command(["markdownhelp", "formatting"], PREFIX_HANDLER)
    & (filters.group | filters.private),
)
async def markdownhelp(_, m: Message):
    await m.reply_text(__help__, quote=True, reply_markup=(await gen_formatting_kb(m)))
    LOGGER.info(f"{m.from_user.id} used cmd '{m.command[0]}' in {m.chat.id}")
    return


@Alita.on_callback_query(filters.regex("^formatting."))
async def get_formatting_info(_, q: CallbackQuery):
    cmd = q.data.split(".")[1]
    kb = InlineKeyboardMarkup(
        [[InlineKeyboardButton("Back", callback_data="back.formatting")]],
    )

    if cmd == "md_formatting":
        await q.message.edit_text(md_formatting_str, reply_markup=kb)
    elif cmd == "fillings":
        await q.message.edit_text(filling_str, reply_markup=kb)
    elif cmd == "random_content":
        await q.message.edit_text(random_content_str, reply_markup=kb)

    await q.answer()
    return


@Alita.on_callback_query(filters.regex("^back."))
async def send_mod_help(_, q: CallbackQuery):
    await q.message.edit_text(
        (tlang(q, "start.private")),
        reply_markup=(await gen_formatting_kb(q.message)),
    )
    await q.answer()
    return


__PLUGIN__ = "plugins.formatting.main"
__help__ = "plugins.formatting.help"
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
