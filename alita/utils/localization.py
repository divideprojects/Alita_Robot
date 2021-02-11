import json
import os.path
from glob import glob
from pyrogram.types import CallbackQuery
from alita.db import lang_db as db
from alita import ENABLED_LOCALES as enabled_locales


def cache_localizations(files):
    ldict = {lang: {} for lang in enabled_locales}
    for file in files:
        lname = file.split(os.path.sep)[1]
        dic = json.load(open(file, encoding="utf-8"))
        ldict[lname].update(dic)
    return ldict


jsons = []
for locale in enabled_locales:
    jsons += glob(os.path.join("locales", locale, "*.json"))


langdict = cache_localizations(jsons)


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

    async def strs(self, string):
        return self.dic.get(string) or langdict["en-US"].get(string) or string
