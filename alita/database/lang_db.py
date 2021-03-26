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


from threading import RLock
from time import time

from alita import LOGGER
from alita.database import MongoDB

INSERTION_LOCK = RLock()

LANG_DATA = {}


class Langs:
    """Class for language options in bot."""

    def __init__(self) -> None:
        self.collection = MongoDB("langs")

    def get_chat_type(self, chat_id: int):
        _ = self
        if str(chat_id).startswith("-100"):
            chat_type = "supergroup"
        else:
            chat_type = "user"
        return chat_type

    def set_lang(self, chat_id: int, lang):
        with INSERTION_LOCK:

            global LANG_DATA
            chat_type = self.get_chat_type(chat_id)
            if chat_id in list(LANG_DATA.keys()):
                try:
                    lang_dict = (LANG_DATA[chat_id]).update({"lang": lang})
                    (LANG_DATA[chat_id]).update(lang_dict)
                except Exception:
                    pass

            curr = self.collection.find_one({"_id": chat_id})
            if curr:
                self.collection.update(
                    {"_id": chat_id},
                    {"lang": lang},
                )
                return "Updated language"

            LANG_DATA[chat_id] = {"chat_type": chat_type, "lang": lang}
            return self.collection.insert_one(
                {"_id": chat_id, "chat_type": chat_type, "lang": lang},
            )

    def get_lang(self, chat_id: int):
        with INSERTION_LOCK:

            global LANG_DATA
            chat_type = self.get_chat_type(chat_id)

            try:
                lang_dict = LANG_DATA[chat_id]
                if lang_dict:
                    user_lang = lang_dict["lang"]
                    return user_lang
            except Exception:
                pass

            curr_lang = self.collection.find_one({"_id": chat_id})
            if curr_lang:
                return curr_lang["lang"]

            LANG_DATA[chat_id] = {"chat_type": chat_type, "lang": "en"}
            self.collection.insert_one(
                {"_id": chat_id, "chat_type": chat_type, "lang": "en"},
            )
            return "en"  # default lang

    def get_all_langs(self):
        return self.collection.find_all()

    # Migrate if chat id changes!
    def migrate_chat(self, old_chat_id: int, new_chat_id: int):
        global LANG_DATA
        with INSERTION_LOCK:

            try:
                old_chat_local = self.get_grp(chat_id=old_chat_id)
                if old_chat_local:
                    lang_dict = LANG_DATA[old_chat_id]
                    del LANG_DATA[old_chat_id]
                    LANG_DATA[new_chat_id] = lang_dict
            except KeyError:
                pass

            old_chat_db = self.collection.find_one({"_id": old_chat_id})
            if old_chat_db:
                new_data = old_chat_db.update({"_id": new_chat_id})
                self.collection.delete_one({"_id": old_chat_id})
                self.collection.insert_one(new_data)


def __load_all_langs():
    global LANG_DATA
    start = time()
    db = Langs()
    langs_data = db.get_all_langs()
    LANG_DATA = {
        int(chat["_id"]): {"lang": chat["lang"], "chat_type": chat["chat_type"]}
        for chat in langs_data
    }
    LOGGER.info(f"Loaded Lang Cache - {round((time()-start),3)}s")
