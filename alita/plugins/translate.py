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
from googletrans import Translator, LANGUAGES
from pyrogram import filters
from pyrogram.types import Message
from alita import PREFIX_HANDLER
from alita.bot_class import Alita
from alita.utils.localization import GetLang


__PLUGIN__ = "TRANSLATE"

__help__ = """
 × /tr <language>: Translates the text and then replies to you with the language \
 you have specifed, works as a reply to message.
"""


def get_lang(text):
    if len(text.split()) > 0:
        lang = text.split()[0]
        if lang.split("-")[0] not in LANGUAGES:
            lang = "pt"
        if len(lang.split("-")) > 1 and lang.split("-")[1] not in LANGUAGES:
            lang = "pt"
    else:
        lang = "pt"
    return lang


@Alita.on_message(filters.command("tr", PREFIX_HANDLER))
async def translate(c: Alita, m: Message):
    _ = GetLang(m).strs
    translator = Translator()
    text = m.text[4:]
    lang = get_lang(text)
    if m.reply_to_message:
        text = m.reply_to_message.text or m.reply_to_message.caption
    else:
        text = text.replace(lang, "", 1).strip() if text.startswith(lang) else text

    if text:
        sent = await m.reply_text(
            _("translate.translating"), reply_to_message_id=m.message_id
        )
        langs = {}

        if len(lang.split("-")) > 1:
            langs["src"] = lang.split("-")[0]
            langs["dest"] = lang.split("-")[1]
        else:
            langs["dest"] = lang

        trres = translator.translate(text, **langs)
        text = trres.text

        res = escape(text)
        await sent.edit_text(
            _("translate.translation").format(
                from_lang=trres.src, to_lang=trres.dest, translation=res
            ),
            parse_mode="HTML",
        )

    else:
        await m.reply_text(
            _("translate.translate_usage"),
            reply_to_message_id=m.message_id,
            parse_mode="markdown",
        )

    return
