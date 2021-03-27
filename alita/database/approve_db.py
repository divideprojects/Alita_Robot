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


class Approve:
    """Class for managing Approves in Chats in Bot."""

    # Database name to connect to to preform operations
    db_name = "approve"

    def __init__(self, chat_id: int) -> None:
        self.collection = MongoDB(self.db_name)
        self.chat_id = chat_id
        self.chat_info = self.__ensure_in_db()

    def check_approve(self, user_id: int):
        with INSERTION_LOCK:
            chat_approved = self.chat_info["users"]
            return bool(user_id in chat_approved.keys())

    def add_approve(self, user_id: int, user_name: str):
        with INSERTION_LOCK:
            new_user_data = {user_id: user_name}
            if not self.check_approve(user_id):
                return self.collection.update(
                    {"_id": self.chat_id},
                    {"users": self.chat_info["users"] | new_user_data},
                )
            return True

    def remove_approve(self, user_id: int):
        with INSERTION_LOCK:
            if self.check_approve(user_id):
                users = self.chat_info["users"].pop(user_id)
                return self.collection.update({"_id": self.chat_id}, {"users": users})
            return True

    def unapprove_all(self):
        with INSERTION_LOCK:
            return self.collection.delete_one(
                {"_id": self.chat_id},
            )

    def list_approved(self):
        with INSERTION_LOCK:
            return self.chat_info["users"].items()

    def count_all_approved(self):
        with INSERTION_LOCK:
            curr = self.collection.find_all()
            return sum([len(set(chat["users"].keys())) for chat in curr])

    def count_approved_chats(self):
        with INSERTION_LOCK:
            return (self.collection.count()) or 0

    def count_approved(self):
        with INSERTION_LOCK:
            return len(self.chat_info["users"])

    def load_from_db(self):
        return self.collection.find_all()

    # Migrate if chat id changes!
    def migrate_chat(self, new_chat_id: int):
        old_chat_db = self.collection.find_one({"_id": self.chat_id})
        new_data = old_chat_db.update({"_id": new_chat_id})
        self.collection.delete_one({"_id": self.chat_id})
        self.collection.insert_one(new_data)

    def __ensure_in_db(self):
        chat_data = self.collection.find_one({"_id": self.chat_id})
        if not chat_data:
            new_data = {"_id": self.chat_id, "users": {}}
            self.collection.insert_one(new_data)
            LOGGER.info(f"Initialized Pins Document for chat {self.chat_id}")
            return new_data
        return chat_data
