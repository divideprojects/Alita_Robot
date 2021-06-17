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
from pyrogram.types import CallbackQuery, Message
from pyromod.helpers import ikb

from alita import LOGGER
from alita.bot_class import Alita
from alita.database.lang_db import Langs
from alita.tr_engine import lang_dict, tlang
from alita.utils.custom_filters import admin_filter, command


async def gen_langs_kb():
    langs = sorted(list(lang_dict.keys()))
    return [
        [
            (
                f"{lang_dict[lang]['main']['language_flag']} {lang_dict[lang]['main']['language_name']} ({lang_dict[lang]['main']['lang_sample']})",
                f"set_lang.{lang}",
            )
            for lang in langs
        ],
        [
            (
                "🌎 Help us with translations!",
                "https://crowdin.com/project/alita_robot",
                "url",
            )
        ],
    ]


@Alita.on_callback_query(filters.regex("^chlang$"))
async def chlang_callback(_, q: CallbackQuery):
    kb = await gen_langs_kb()
    kb.append([(f"« {(tlang(q, 'general.back_btn'))}", "start_back")])

    await q.message.edit_text(
        (tlang(q, "langs.changelang")),
        reply_markup=ikb(kb),
    )
    await q.answer()
    return


@Alita.on_callback_query(filters.regex("^close$"), group=3)
async def close_btn_callback(_, q: CallbackQuery):
    await q.message.delete()
    try:
        await q.message.reply_to_message.delete()
    except Exception as ef:
        LOGGER.error(f"Error: Cannot delete message\n{ef}")
    await q.answer()
    return


@Alita.on_callback_query(filters.regex("^set_lang."))
async def set_lang_callback(_, q: CallbackQuery):
    lang_code = q.data.split(".")[1]

    Langs(q.message.chat.id).set_lang(lang_code)
    await sleep(0.1)

    if q.message.chat.type == "private":
        keyboard = ikb([[(f"« {(tlang(q, 'general.back_btn'))}", "start_back")]])
    else:
        keyboard = None
    await q.message.edit_text(
        f"🌐 {((tlang(q, 'langs.changed')).format(lang_code=lang_code))}",
        reply_markup=keyboard,
    )
    await q.answer()
    return


@Alita.on_message(
    command(["lang", "setlang"]) & (admin_filter | filters.private),
    group=7,
)
async def set_lang(_, m: Message):
    args = m.text.split()

    if len(args) > 2:
        await m.reply_text(tlang(m, "langs.correct_usage"))
        return
    if len(args) == 2:
        lang_code = args[1]
        avail_langs = set(lang_dict.keys())
        if lang_code not in avail_langs:
            await m.reply_text(
                f"Please choose a valid language code from: {', '.join(avail_langs)}",
            )
            return
        Langs(m.chat.id).set_lang(lang_code)
        LOGGER.info(f"{m.from_user.id} change language to {lang_code} in {m.chat.id}")
        await m.reply_text(
            f"🌐 {((tlang(m, 'langs.changed')).format(lang_code=lang_code))}",
        )
        return
    await m.reply_text(
        (tlang(m, "langs.changelang")),
        reply_markup=ikb(await gen_langs_kb()),
    )
    return


__PLUGIN__ = "language"

__alt_name__ = ["lang", "langs", "languages"]
__buttons__ = [
    [
        (
            "🌎 Help us with translations!",
            "https://crowdin.com/project/alita_robot",
            "url",
        ),
    ],
]
