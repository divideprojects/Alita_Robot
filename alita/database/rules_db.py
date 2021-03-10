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


class Rules:
    """Class for rules for chats in bot."""

    def __init__(self) -> None:
        self.collection = MongoDB("rules")

    def get_rules(self, chat_id: int):
        with INSERTION_LOCK:
            rules = self.collection.find_one({"_id": chat_id})
            if rules:
                return rules["rules"]
            return None

    def set_rules(self, chat_id: int, rules: str, privrules: bool = False):
        with INSERTION_LOCK:
            curr_rules = self.collection.find_one({"_id": chat_id})
            if curr_rules:
                return self.collection.update(
                    {"_id": chat_id},
                    {"rules": rules},
                )
            return self.collection.insert_one(
                {"_id": chat_id, "rules": rules, "privrules": privrules},
            )

    def set_privrules(self, chat_id: int, privrules: bool):
        with INSERTION_LOCK:
            curr_rules = self.collection.find_one({"_id": chat_id})
            if curr_rules:
                return self.collection.update(
                    {"_id": chat_id},
                    {"privrules": privrules},
                )
            return self.collection.insert_one({"_id": chat_id, "privrules": privrules})

    def get_privrules(self, chat_id: int):
        with INSERTION_LOCK:
            curr_rules = self.collection.find_one({"_id": chat_id})
            if curr_rules:
                return curr_rules["privrules"]
            return self.collection.insert_one({"_id": chat_id, "privrules": False})

    def clear_rules(self, chat_id: int):
        with INSERTION_LOCK:
            curr_rules = self.collection.find_one({"_id": chat_id})
            if curr_rules:
                return self.collection.delete_one({"_id": chat_id})
            return

    def count_chats(self):
        with INSERTION_LOCK:
            return self.collection.count({"rules": {"$regex": ".*"}})

    def count_privrules_chats(self):
        with INSERTION_LOCK:
            return self.collection.count({"privrules": True})

    def count_grouprules_chats(self):
        with INSERTION_LOCK:
            return self.collection.count({"privrules": False})

    # Migrate if chat id changes!
    def migrate_chat(self, old_chat_id: int, new_chat_id: int):
        with INSERTION_LOCK:

            old_chat_db = self.collection.find_one({"_id": old_chat_id})
            if old_chat_db:
                new_data = old_chat_db.update({"_id": new_chat_id})
                self.collection.delete_one({"_id": old_chat_id})
                self.collection.insert_one(new_data)
            return
