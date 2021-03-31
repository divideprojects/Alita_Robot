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

from alita import LOGGER
from alita.database import MongoDB

INSERTION_LOCK = RLock()


class Blacklist:
    """Class to manage database for blacklists for chats."""

    # Database name to connect to to preform operations
    db_name = "blacklists"

    def __init__(self, chat_id: int) -> None:
        self.collection = MongoDB(self.db_name)
        self.chat_id = chat_id
        self.chat_info = self.__ensure_in_db()

    def check_word_blacklist_status(self, word: str):
        with INSERTION_LOCK:
            bl_words = self.chat_info["triggers"]
            return bool(word in bl_words)

    def add_blacklist(self, trigger: str):
        with INSERTION_LOCK:
            if not self.check_word_blacklist_status(trigger):
                return self.collection.update(
                    {"_id": self.chat_id},
                    {
                        "_id": self.chat_id,
                        "triggers": self.chat_info["triggers"] + [trigger],
                    },
                )

    def remove_blacklist(self, trigger: str):
        with INSERTION_LOCK:
            if self.check_word_blacklist_status(trigger):
                self.chat_info["triggers"].remove(trigger)
                return self.collection.update(
                    {"_id": self.chat_id},
                    {
                        "_id": self.chat_id,
                        "triggers": self.chat_info["triggers"],
                    },
                )

    def get_blacklists(self):
        with INSERTION_LOCK:
            return self.chat_info["triggers"]

    @staticmethod
    def count_blacklists_all():
        with INSERTION_LOCK:
            collection = MongoDB(Blacklist.db_name)
            curr = collection.find_all()
            return sum(len(chat["triggers"]) for chat in curr)

    @staticmethod
    def count_blackists_chats():
        with INSERTION_LOCK:
            collection = MongoDB(Blacklist.db_name)
            curr = collection.find_all()
            return sum([1 for chat in curr if chat["triggers"]])

    def set_action(self, action: str):
        with INSERTION_LOCK:
            return self.collection.update(
                {"_id": self.chat_id},
                {"_id": self.chat_id, "action": action},
            )

    def get_action(self):
        with INSERTION_LOCK:
            return self.chat_info["action"]

    def set_reason(self, reason: str):
        with INSERTION_LOCK:
            return self.collection.update(
                {"_id": self.chat_id},
                {"_id": self.chat_id, "reason": reason},
            )

    def get_reason(self):
        with INSERTION_LOCK:
            return self.chat_info["reason"]

    @staticmethod
    def count_action_bl_all(action: str):
        with INSERTION_LOCK:
            collection = MongoDB(Blacklist.db_name)
            return collection.count({"action": action})

    def rm_all_blacklist(self):
        with INSERTION_LOCK:
            return self.collection.update(
                {"_id": self.chat_id},
                {"triggers": []},
            )

    def __ensure_in_db(self):
        chat_data = self.collection.find_one({"_id": self.chat_id})
        if not chat_data:
            new_data = new_data = {
                "_id": self.chat_id,
                "triggers": [],
                "action": "none",
                "reason": "Automated blacklisted word",
            }
            self.collection.insert_one(new_data)
            LOGGER.info(f"Initialized Blacklist Document for chat {self.chat_id}")
            return new_data
        return chat_data

    # Migrate if chat id changes!
    def migrate_chat(self, new_chat_id: int):
        old_chat_db = self.collection.find_one({"_id": self.chat_id})
        new_data = old_chat_db.update({"_id": new_chat_id})
        self.collection.insert_one(new_data)
        self.collection.delete_one({"_id": self.chat_id})

    @staticmethod
    def repair_db(collection):
        all_data = collection.find_all()
        keys = {
            "triggers": [],
            "action": "none",
            "reason": "Automated blacklisted word",
        }
        for data in all_data:
            for key, val in keys.items():
                try:
                    _ = data[key]
                except KeyError:
                    LOGGER.warning(
                        f"Repairing Blacklist Database - setting '{key}:{val}' for {data['_id']}",
                    )
                    collection.update({"_id": data["_id"]}, {key: val})


def __check_db_status():
    collection = MongoDB(Blacklist.db_name)
    Blacklist.repair_db(collection)


__check_db_status()
