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

# Locall cache languages for users!!
LANG_CACHE = {}


class Langs(MongoDB):
    """Class for language options in bot."""

    db_name = "langs"

    def __init__(self, chat_id: int) -> None:
        super().__init__(self.db_name)
        self.chat_id = chat_id
        self.chat_info = self.__ensure_in_db()

    def get_chat_type(self):
        return "supergroup" if str(self.chat_id).startswith("-100") else "user"

    def set_lang(self, lang: str):
        with INSERTION_LOCK:
            global LANG_CACHE
            LANG_CACHE[self.chat_id] = lang
            self.chat_info["lang"] = lang
            return self.update(
                {"_id": self.chat_id},
                {"lang": self.chat_info["lang"]},
            )

    def get_lang(self):
        with INSERTION_LOCK:
            return self.chat_info["lang"]

    @staticmethod
    def load_from_db():
        with INSERTION_LOCK:
            collection = MongoDB(Langs.db_name)
            return collection.find_all()

    def __ensure_in_db(self):
        try:
            chat_data = {"_id": self.chat_id, "lang": LANG_CACHE[self.chat_id]}
        except KeyError:
            chat_data = self.find_one({"_id": self.chat_id})
        if not chat_data:
            chat_type = self.get_chat_type()
            new_data = {"_id": self.chat_id, "lang": "en", "chat_type": chat_type}
            self.insert_one(new_data)
            LOGGER.info(f"Initialized Language Document for chat {self.chat_id}")
            return new_data
        return chat_data

    # Migrate if chat id changes!
    def migrate_chat(self, new_chat_id: int):
        old_chat_db = self.find_one({"_id": self.chat_id})
        new_data = old_chat_db.update({"_id": new_chat_id})
        self.insert_one(new_data)
        self.delete_one({"_id": self.chat_id})

    @staticmethod
    def repair_db(collection):
        all_data = collection.find_all()
        keys = {"lang": "en", "chat_type": ""}
        for data in all_data:
            for key, val in keys.items():
                try:
                    _ = data[key]
                except KeyError:
                    LOGGER.warning(
                        f"Repairing Langs Database - setting '{key}:{val}' for {data['_id']}",
                    )
                    collection.update({"_id": data["_id"]}, {key: val})


def __pre_req_all_langs():
    start = time()
    LOGGER.info("Starting Langs Database Repair...")
    collection = MongoDB(Langs.db_name)
    Langs.repair_db(collection)
    LOGGER.info(f"Done in {round((time() - start), 3)}s!")


def __load_lang_cache():
    global LANG_CACHE
    collection = MongoDB(Langs.db_name)
    all_data = collection.find_all()
    LANG_CACHE = {i["_id"]: i["lang"] for i in all_data}
