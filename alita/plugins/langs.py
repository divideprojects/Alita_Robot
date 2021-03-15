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

from pyrogram import filters
from pyrogram.types import (
    CallbackQuery,
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    Message,
)

from alita import PREFIX_HANDLER
from alita.bot_class import Alita
from alita.database.lang_db import Langs
from alita.tr_engine import lang_dict, tlang
from alita.utils.custom_filters import admin_filter

# initialise
db = Langs()

__PLUGIN__ = "plugins.language.main"
__help__ = "plugins.language.help"


async def gen_langs_kb():
    langs = list(lang_dict.keys())
    kb = []
    while langs:
        lang_main = lang_dict[langs[0]]["main"]
        a = [
            InlineKeyboardButton(
                f"{lang_main['language_flag']} {lang_main['language_name']} ({lang_main['lang_sample']})",
                callback_data=f"set_lang.{langs[0]}",
            ),
        ]
        langs.pop(0)
        if langs:
            lang_main = lang_dict[langs[0]]["main"]
            a.append(
                InlineKeyboardButton(
                    f"{lang_main['language_flag']} {lang_main['language_name']} ({lang_main['lang_sample']})",
                    callback_data=f"set_lang.{langs[0]}",
                ),
            )
            langs.pop(0)
        kb.append(a)
    kb.append(
        [
            InlineKeyboardButton(
                "🌎 Help us with translations!",
                url="https://crowdin.com/project/alita_robot",
            ),
        ],
    )
    return kb


@Alita.on_callback_query(filters.regex("^chlang$"))
async def chlang_callback(_, q: CallbackQuery):

    keyboard = InlineKeyboardMarkup(
        inline_keyboard=[
            *(await gen_langs_kb()),
            [
                InlineKeyboardButton(
                    f"« {(tlang(q, 'general.back_btn'))}",
                    callback_data="start_back",
                ),
            ],
        ],
    )
    await q.message.edit_text(
        (tlang(q, "langs.changelang")),
        reply_markup=keyboard,
    )
    await q.answer()
    return


@Alita.on_callback_query(filters.regex("^close$"))
async def close_btn_callback(_, q: CallbackQuery):
    await q.message.delete()
    await q.answer()
    return


@Alita.on_callback_query(filters.regex("^set_lang."))
async def set_lang_callback(_, q: CallbackQuery):

    lang_code = q.data.split(".")[1]

    db.set_lang(q.message.chat.id, lang_code)
    await sleep(0.1)

    if q.message.chat.type == "private":
        keyboard = InlineKeyboardMarkup(
            inline_keyboard=[
                [
                    InlineKeyboardButton(
                        f"« {(tlang(q, 'general.back_btn'))}",
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
                        f"❌ {(tlang(q, 'general.close_btn'))}",
                        callback_data="close",
                    ),
                ],
            ],
        )
    await q.message.edit_text(
        f"🌐 {((tlang(q, 'langs.changed')).format(lang_code=lang_code))}",
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
        (tlang(m, "langs.changelang")),
        reply_markup=InlineKeyboardMarkup([*(await gen_langs_kb())]),
    )
    return
