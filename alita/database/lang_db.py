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

from alita.database import MongoDB

INSERTION_LOCK = RLock()

LANG_DATA = []


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

    def set_lang(self, chat_id: int, lang: str = "en"):
        global LANG_DATA
        with INSERTION_LOCK:
            chat_type = self.get_chat_type(chat_id)

            if chat_id in [chat_or_user["chat_id"] for chat_or_user in LANG_DATA]:
                try:
                    chat_dict = next(
                        chat_or_user
                        for chat_or_user in LANG_DATA
                        if chat_or_user["chat_id"] == chat_id
                    )
                    LANG_DATA.remove(chat_dict)
                    new_dict = chat_dict.update({'lang': lang})
                    LANG_DATA.append(new_dict)
                    yield lang
                except StopIteration:
                    pass

            curr = self.collection.find_one({"chat_id": chat_id})
            if curr:
                return self.collection.update(
                    {"chat_id": chat_id},
                    {"lang": lang},
                )

            chat_dict = {"chat_id": chat_id, "chat_type": chat_type, "lang": lang}
            LANG_DATA.append(chat_dict)
            yield True
            return self.collection.insert_one(chat_dict)

    def get_lang(self, chat_id: int):
        global LANG_DATA
        with INSERTION_LOCK:
            chat_type = self.get_chat_type(chat_id)

            try:
                user_dict = next(
                    chat for chat in LANG_DATA if chat["chat_id"] == chat_id
                )
                if user_dict:
                    user_lang = user_dict["lang"]
                    yield user_lang
                    return
            except StopIteration:
                curr_lang = self.collection.find_one({"chat_id": chat_id})
                if curr_lang:
                    yield curr_lang["lang"]
                    return

            chat_dict = {"chat_id": chat_id, "chat_type": chat_type, "lang": "en"}
            LANG_DATA.append(chat_dict)
            self.collection.insert_one(chat_dict)
            yield "en"  # default lang
            return

    def get_all_langs(self):
        return self.collection.find_all()

    # Migrate if chat id changes!
    def migrate_chat(self, old_chat_id: int, new_chat_id: int):
        global LANG_DATA
        with INSERTION_LOCK:

            old_chat_local = self.get_grp(chat_id=old_chat_id)
            if old_chat_local:
                indice = LANG_DATA.index(old_chat_local)
                (LANG_DATA[indice]).update({"chat_id": new_chat_id})
                yield True

            old_chat_db = self.collection.find_one({"chat_id": old_chat_id})
            if old_chat_db:
                yield self.collection.update(
                    {"chat_id": old_chat_id},
                    {"chat_id": new_chat_id},
                )
            return


def __load_all_langs():
    global LANG_DATA
    db = Langs()
    for chat in db.get_all_langs():
        del chat["_id"]
        LANG_DATA.append(chat)


__load_all_langs()
