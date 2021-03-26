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
from traceback import format_exc

from pymongo.errors import DuplicateKeyError

from alita import LOGGER
from alita.database import MongoDB

INSERTION_LOCK = RLock()

PINS_CACHE = {}


class Pins:
    """Class for managing antichannelpins in chats."""

    def __init__(self) -> None:
        self.collection = MongoDB("antichannelpin")

    def check_status(self, chat_id: int, atype: str):
        with INSERTION_LOCK:

            if chat_id in (PINS_CACHE[atype]):
                return True

            curr = self.collection.find_one({"_id": chat_id, atype: True})
            if curr:
                return True

            return False

    def get_current_stngs(self, chat_id: int):
        with INSERTION_LOCK:
            curr = self.collection.find_one({"_id": chat_id})
            if curr:
                return curr

            curr = {"_id": chat_id, "antichannelpin": False, "cleanlinked": False}
            self.collection.insert_one(curr)
            return curr

    def set_on(self, chat_id: int, atype: str):
        global PINS_CACHE
        with INSERTION_LOCK:
            otype = "cleanlinked" if atype == "antichannelpin" else "antichannelpin"
            if chat_id not in (PINS_CACHE[atype]):
                (PINS_CACHE[atype]).add(chat_id)
                try:
                    return self.collection.insert_one(
                        {"_id": chat_id, atype: True, otype: False},
                    )
                except DuplicateKeyError:
                    return self.collection.update(
                        {"_id": chat_id},
                        {atype: True, otype: False},
                    )
            return "Already exists"

    def set_off(self, chat_id: int, atype: str):
        global PINS_CACHE
        with INSERTION_LOCK:
            if chat_id in (PINS_CACHE[atype]):
                (PINS_CACHE[atype]).remove(chat_id)
                return self.collection.update({"_id": chat_id}, {atype: False})
            return f"{atype} not enabled"

    def count_chats(self, atype):
        with INSERTION_LOCK:
            try:
                return len(PINS_CACHE[atype])
            except Exception as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())
                return self.collection.count({atype: True})

    def load_chats_from_db(self, query=None):
        with INSERTION_LOCK:
            if query is None:
                query = {}
            return self.collection.find_all(query)

    def list_chats(self, query):
        with INSERTION_LOCK:
            try:
                return PINS_CACHE[query]
            except Exception as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())
                return self.collection.find_all({query: True})

    # Migrate if chat id changes!
    def migrate_chat(self, old_chat_id: int, new_chat_id: int):
        global PINS_CACHE
        with INSERTION_LOCK:

            # Update locally
            if old_chat_id in (PINS_CACHE["antichannelpin"]):
                (PINS_CACHE["antichannelpin"]).remove(old_chat_id)
                (PINS_CACHE["antichannelpin"]).add(new_chat_id)
            if old_chat_id in (PINS_CACHE["cleanlinked"]):
                (PINS_CACHE["cleanlinked"]).remove(old_chat_id)
                (PINS_CACHE["cleanlinked"]).add(new_chat_id)

            old_chat_db = self.collection.find_one({"_id": old_chat_id})
            if old_chat_db:
                new_data = old_chat_db.update({"_id": new_chat_id})
                self.collection.delete_one({"_id": old_chat_id})
                self.collection.insert_one(new_data)


def __load_pins_chats():
    global PINS_CACHE
    start = time()
    db = Pins()
    all_chats = db.load_chats_from_db()
    PINS_CACHE["antichannelpin"] = {i["_id"] for i in all_chats if i["antichannelpin"]}
    PINS_CACHE["cleanlinked"] = {i["_id"] for i in all_chats if i["cleanlinked"]}
    LOGGER.info(f"Loaded Pins Cache - {round((time()-start),3)}s")
