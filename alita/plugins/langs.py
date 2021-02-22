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


from asyncio import sleep

import lang_db as db
from pyrogram import filters
from pyrogram.types import (
    CallbackQuery,
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    Message,
)

from alita import PREFIX_HANDLER
from alita.bot_class import Alita
from alita.tr_engine import lang_dict, tlang
from alita.utils.custom_filters import admin_filter

__PLUGIN__ = "Language"

__help__ = """
Not able to change language of the bot?
Easily change by using this module!

Just type /lang and use inline keyboard to choose a language \
for yourself or your group.
"""


async def gen_langs_kb():
    langs = list(lang_dict.keys())
    kb = []
    while langs:
        lang = lang_dict[langs[0]]["main"]
        a = [
            InlineKeyboardButton(
                f"{lang['language_flag']} {lang['language_name']} ({lang['lang_sample']})",
                callback_data=f"set_lang.{langs[0]}",
            ),
        ]
        langs.pop(0)
        if langs:
            lang = lang_dict[langs[0]]
            a.append(
                InlineKeyboardButton(
                    f"{lang['language_flag']} {lang['language_name']} ({lang['lang_sample']})",
                    callback_data=f"set_lang.{langs[0]}",
                ),
            )
            langs.pop(0)
        kb.append(a)
    return kb


@Alita.on_callback_query(filters.regex("^chlang$"))
async def chlang_callback(_, q: CallbackQuery):

    keyboard = InlineKeyboardMarkup(
        inline_keyboard=[
            *(await gen_langs_kb()),
            [
                InlineKeyboardButton(
                    f"« {tlang(q, 'general.back_btn')}",
                    callback_data="start_back",
                ),
            ],
        ],
    )
    await q.message.edit_text(tlang(q, "langs.changelang"), reply_markup=keyboard)
    await q.answer()
    return


@Alita.on_callback_query(filters.regex("^close$"))
async def close_btn_callback(_, q: CallbackQuery):
    await q.message.delete()
    await q.answer()
    return


@Alita.on_callback_query(filters.regex("^set_lang."))
async def set_lang_callback(_, q: CallbackQuery):

    db.set_lang(q.message.chat.id, q.message.chat.type, q.data.split(".")[1])
    await sleep(0.1)

    if q.message.chat.type == "private":
        keyboard = InlineKeyboardMarkup(
            inline_keyboard=[
                [
                    InlineKeyboardButton(
                        f"« {tlang(q, 'general.back_btn')}",
                        callback_data="start_back",
                    ),
                ],
            ],
        )
    else:
        keyboard = InlineKeyboardMarkup(
            inline_keyboard=[
                [
                    InlineKeyboardButton(
                        f"❌ {tlang(q, 'general.close_btn')}",
                        callback_data="close",
                    ),
                ],
            ],
        )
    lang_code = q.data.split(".")[1]
    await q.message.edit_text(
        f"🌐 {tlang(q, 'langs.changed').format(lang_code=lang_code)}",
        reply_markup=keyboard,
    )
    await q.answer()
    return


@Alita.on_message(
    filters.command(["lang", "setlang"], PREFIX_HANDLER)
    & ((admin_filter & filters.group) | filters.private),
)
async def set_lang(_, m: Message):

    if len(m.text.split()) >= 2:
        await m.reply_text(tlang(m, "langs.correct_usage"))
        return
    await m.reply_text(
        tlang(m, "langs.changelang"),
        reply_markup=InlineKeyboardMarkup([*(await gen_langs_kb())]),
    )
    return
