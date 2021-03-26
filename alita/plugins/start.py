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
from pyrogram.errors import MessageNotModified, QueryIdInvalid, UserIsBlocked
from pyrogram.types import (
    CallbackQuery,
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    Message,
)

from alita import HELP_COMMANDS, LOGGER, PREFIX_HANDLER
from alita.bot_class import Alita
from alita.tr_engine import tlang
from alita.utils.start_utils import (
    gen_cmds_kb,
    gen_start_kb,
    get_help_msg,
    get_private_note,
    get_private_rules,
)


@Alita.on_message(
    filters.command("donate", PREFIX_HANDLER) & (filters.group | filters.private),
)
async def donate(_, m: Message):
    LOGGER.info(f"{m.from_user.id} fetched donation text in {m.chat.id}")
    await m.reply_text(tlang(m, "general.donate_owner"))
    return


@Alita.on_message(
    filters.command("start", PREFIX_HANDLER) & (filters.group | filters.private),
)
async def start(c: Alita, m: Message):

    if m.chat.type == "private":
        if len(m.text.split()) > 1:
            help_option = (m.text.split(None, 1)[1]).lower()

            if help_option.startswith("note"):
                await get_private_note(c, m, help_option)
                return
            if help_option.startswith("rules"):
                LOGGER.info(f"{m.from_user.id} fetched privaterules in {m.chat.id}")
                await get_private_rules(c, m, help_option)
                return

            help_msg, help_kb = await get_help_msg(m, help_option)

            if not help_msg:
                return

            await m.reply_text(
                help_msg,
                parse_mode="markdown",
                reply_markup=InlineKeyboardMarkup(help_kb),
                quote=True,
                disable_web_page_preview=True,
            )
            return
        try:
            await m.reply_text(
                (tlang(m, "start.private")),
                reply_markup=(await gen_start_kb(m)),
                quote=True,
                disable_web_page_preview=True,
            )
        except UserIsBlocked:
            LOGGER.warning(f"Bot blocked by {m.from_user.id}")
    else:
        await m.reply_text(
            (tlang(m, "start.group")),
            quote=True,
        )
    return


@Alita.on_callback_query(filters.regex("^start_back$"))
async def start_back(_, q: CallbackQuery):

    try:
        await q.message.edit_text(
            (tlang(q, "start.private")),
            reply_markup=(await gen_start_kb(q.message)),
            disable_web_page_preview=True,
        )
    except MessageNotModified:
        pass
    await q.answer()
    return


@Alita.on_callback_query(filters.regex("^commands$"))
async def commands_menu(_, q: CallbackQuery):

    keyboard = InlineKeyboardMarkup(
        inline_keyboard=[
            *(await gen_cmds_kb(q)),
            [
                InlineKeyboardButton(
                    f"« {(tlang(q, 'general.back_btn'))}",
                    callback_data="start_back",
                ),
            ],
        ],
    )
    try:
        await q.message.edit_text(
            (tlang(q, "general.commands_available")),
            reply_markup=keyboard,
        )
    except MessageNotModified:
        pass
    except QueryIdInvalid:
        await q.message.reply_text(
            (tlang(q, "general.commands_available")),
            reply_markup=keyboard,
        )
    await q.answer()
    return


@Alita.on_message(filters.command("help", PREFIX_HANDLER))
async def help_menu(_, m: Message):

    from alita import BOT_USERNAME

    if len(m.text.split()) >= 2:
        help_option = (m.text.split(None, 1)[1]).lower()
        help_msg, help_kb = await get_help_msg(m, help_option)

        if not help_msg:
            LOGGER.error(f"No help_msg found for help_option - {help_option}!!")
            return

        LOGGER.info(
            f"{m.from_user.id} fetched help for '{help_option}' text in {m.chat.id}",
        )
        if m.chat.type == "private":
            await m.reply_text(
                help_msg,
                parse_mode="markdown",
                reply_markup=InlineKeyboardMarkup(help_kb),
                quote=True,
                disable_web_page_preview=True,
            )
        else:
            await m.reply_text(
                (tlang(m, "start.public_help").format(help_option=help_option)),
                reply_markup=InlineKeyboardMarkup(
                    [
                        [
                            InlineKeyboardButton(
                                "Help",
                                url=f"t.me/{BOT_USERNAME}?start={help_option}",
                            ),
                        ],
                    ],
                ),
            )
    else:
        if m.chat.type == "private":
            keyboard = InlineKeyboardMarkup(
                [
                    *(await gen_cmds_kb(m)),
                    [
                        InlineKeyboardButton(
                            f"« {(tlang(m, 'general.back_btn'))}",
                            callback_data="start_back",
                        ),
                    ],
                ],
            )
            msg = tlang(m, "general.commands_available")
        else:
            keyboard = InlineKeyboardMarkup(
                [
                    [
                        InlineKeyboardButton(
                            text="Help",
                            url=f"t.me/{BOT_USERNAME}?start=help",
                        ),
                    ],
                ],
            )
            msg = tlang(m, "start.pm_for_help")

        await m.reply_text(
            msg,
            reply_markup=keyboard,
        )

    return


@Alita.on_callback_query(filters.regex("^get_mod."))
async def get_module_info(_, q: CallbackQuery):

    module = q.data.split(".", 1)[1]

    help_msg = f"**{(tlang(q, str(module)))}:**\n\n" + tlang(
        q,
        HELP_COMMANDS[module]["help_msg"],
    )

    help_kb = HELP_COMMANDS[module]["buttons"] + [
        [
            InlineKeyboardButton(
                "« " + (tlang(q, "general.back_btn")),
                callback_data="commands",
            ),
        ],
    ]
    await q.message.edit_text(
        help_msg,
        parse_mode="markdown",
        reply_markup=InlineKeyboardMarkup(help_kb),
    )
    await q.answer()
    return
