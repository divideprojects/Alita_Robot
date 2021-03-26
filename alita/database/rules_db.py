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

from alita import LOGGER
from alita.database import MongoDB

INSERTION_LOCK = RLock()

RULES_CACHE = {}


class Rules:
    """Class for rules for chats in bot."""

    def __init__(self) -> None:
        self.collection = MongoDB("rules")

    def get_rules(self, chat_id: int):
        global RULES_CACHE
        with INSERTION_LOCK:

            if (chat_id in set(RULES_CACHE.keys())) and (RULES_CACHE[chat_id]["rules"]):
                return RULES_CACHE[chat_id]["rules"]

            rules = self.collection.find_one({"_id": chat_id})
            if rules:
                return rules["rules"]
            return None

    def set_rules(self, chat_id: int, rules: str, privrules: bool = False):
        global RULES_CACHE
        with INSERTION_LOCK:

            if chat_id in set(RULES_CACHE.keys()):
                RULES_CACHE[chat_id]["rules"] = rules

            curr_rules = self.collection.find_one({"_id": chat_id})
            if curr_rules:
                return self.collection.update(
                    {"_id": chat_id},
                    {"rules": rules},
                )

            RULES_CACHE[chat_id] = {"rules": rules, "privrules": privrules}
            return self.collection.insert_one(
                {"_id": chat_id, "rules": rules, "privrules": privrules},
            )

    def get_privrules(self, chat_id: int):
        global RULES_CACHE
        with INSERTION_LOCK:

            if (chat_id in set(RULES_CACHE.keys())) and (
                RULES_CACHE[chat_id]["privrules"]
            ):
                return RULES_CACHE[chat_id]["privrules"]

            curr_rules = self.collection.find_one({"_id": chat_id})
            if curr_rules:
                return curr_rules["privrules"]

            RULES_CACHE[chat_id] = {"privrules": False, "rules": ""}
            return self.collection.insert_one(
                {"_id": chat_id, "rules": "", "privrules": False},
            )

    def set_privrules(self, chat_id: int, privrules: bool):
        global RULES_CACHE
        with INSERTION_LOCK:

            if chat_id in set(RULES_CACHE.keys()):
                RULES_CACHE[chat_id]["privrules"] = privrules

            curr_rules = self.collection.find_one({"_id": chat_id})
            if curr_rules:
                return self.collection.update(
                    {"_id": chat_id},
                    {"privrules": privrules},
                )

            RULES_CACHE[chat_id] = {"rules": "", "privrules": privrules}
            return self.collection.insert_one(
                {"_id": chat_id, "rules": "", "privrules": privrules},
            )

    def clear_rules(self, chat_id: int):
        global RULES_CACHE
        with INSERTION_LOCK:

            if chat_id in set(RULES_CACHE.keys()):
                del RULES_CACHE[chat_id]

            curr_rules = self.collection.find_one({"_id": chat_id})
            if curr_rules:
                return self.collection.delete_one({"_id": chat_id})
            return "Rules not found!"

    def count_chats(self):
        with INSERTION_LOCK:
            try:
                return len([i for i in RULES_CACHE if RULES_CACHE[i]["rules"]])
            except Exception as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())
            return self.collection.count({"rules": {"$regex": ".*"}})

    def count_privrules_chats(self):
        with INSERTION_LOCK:
            try:
                return len([i for i in RULES_CACHE if RULES_CACHE[i]["privrules"]])
            except Exception as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())
                return self.collection.count({"privrules": True})

    def count_grouprules_chats(self):
        with INSERTION_LOCK:
            try:
                return len([i for i in RULES_CACHE if not RULES_CACHE[i]["privrules"]])
            except Exception as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())
                return self.collection.count({"privrules": True})

    def load_from_db(self):
        return self.collection.find_all()

    # Migrate if chat id changes!
    def migrate_chat(self, old_chat_id: int, new_chat_id: int):
        global RULES_CACHE
        with INSERTION_LOCK:

            # Update locally
            try:
                old_db_local = RULES_CACHE[old_chat_id]
                del RULES_CACHE[old_chat_id]
                RULES_CACHE[new_chat_id] = old_db_local
            except KeyError:
                pass

            # Update in db
            old_chat_db = self.collection.find_one({"_id": old_chat_id})
            if old_chat_db:
                new_data = old_chat_db.update({"_id": new_chat_id})
                self.collection.delete_one({"_id": old_chat_id})
                self.collection.insert_one(new_data)


def __load_all_rules():
    global RULES_CACHE
    start = time()
    db = Rules()
    data = db.load_from_db()
    RULES_CACHE = {
        int(chat["_id"]): {
            "rules": chat["rules"],
            "privrules": chat["privrules"],
        }
        for chat in data
    }
    LOGGER.info(f"Loaded Rules Cache - {round((time()-start),3)}s")
