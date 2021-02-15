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


from glob import glob
from os import path

from pyrogram.types import CallbackQuery
from ujson import load

from alita import ENABLED_LOCALES as enabled_locales
from alita.db import lang_db as db


async def cache_localizations(files):
    ldict = {lang: {} for lang in enabled_locales}
    for file in files:
        lname = file.split(path.sep)[1]
        dic = load(open(file, encoding="utf-8"))
        ldict[lname].update(dic)
    return ldict


langdict = None


async def load_langdict():
    jsons = []
    for locale in enabled_locales:
        jsons += glob(path.join("locales", locale, "*.json"))
    global langdict
    langdict = await cache_localizations(jsons)
    if langdict:
        return True
    return False


class GetLang:
    def __init__(self, msg):
        if isinstance(msg, CallbackQuery):
            chat = msg.message.chat
        else:
            chat = msg.chat

        lang = db.get_lang(chat.id, chat.type)
        if chat.type == "private":
            self.lang = lang or msg.from_user.language_code or "en-US"
        else:
            self.lang = lang or "en-US"
        # User has a language_code without hyphen
        if len(self.lang.split("-")) == 1:
            # Try to find a language that starts with the provided
            # language_code
            for locale_ in enabled_locales:
                if locale_.startswith(self.lang):
                    self.lang = locale_
        elif self.lang.split("-")[1].islower():
            self.lang = self.lang.split("-")
            self.lang[1] = self.lang[1].upper()
            self.lang = "-".join(self.lang)
        self.lang = self.lang if self.lang in enabled_locales else "en-US"

        self.dic = langdict.get(self.lang, langdict["en-US"])

    def strs(self, string):
        return self.dic.get(string) or langdict["en-US"].get(string) or string
