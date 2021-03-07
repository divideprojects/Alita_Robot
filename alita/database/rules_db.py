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

    async def get_rules(self, chat_id: int):
        with INSERTION_LOCK:
            rules = self.collection.find_one({"chat_id": chat_id})
            if rules:
                return rules["rules"]
            return None

    async def set_rules(self, chat_id: int, rules: str):
        with INSERTION_LOCK:
            curr_rules = self.collection.find_one({"chat_id": chat_id})
            if curr_rules:
                return self.collection.update(
                    {"chat_id": chat_id},
                    {"rules": rules},
                )
            return self.collection.insert_one({"chat_id": chat_id, "rules": rules})

    async def clear_rules(self, chat_id: int):
        with INSERTION_LOCK:
            curr_rules = self.collection.find_one({"chat_id": chat_id})
            if curr_rules:
                return self.collection.delete_one({"chat_id": chat_id})
            return

    async def count_chats(self):
        with INSERTION_LOCK:
            return self.collection.count()

    # Migrate if chat id changes!
    async def migrate_chat(self, old_chat_id: int, new_chat_id: int):
        with INSERTION_LOCK:
            old_chat = self.collection.find_one({"chat_id": old_chat_id})
            if old_chat:
                return self.collection.update(
                    {"chat_id": old_chat_id},
                    {"chat_id": new_chat_id},
                )
            return
