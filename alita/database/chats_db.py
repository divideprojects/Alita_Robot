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


class Chats:
    """Class to manage users for bot."""

    def __init__(self) -> None:
        self.collection = MongoDB("chats")

    def remove_chat(self, chat_id: int):
        with INSERTION_LOCK:
            self.collection.delete_one({"_id": chat_id})

    def update_chat(self, chat_id: int, chat_name: str, user_id: int):
        with INSERTION_LOCK:
            curr = self.collection.find_one({"_id": chat_id})
            if curr:
                users_old = curr["users"]
                users_old.append(user_id)
                users = list(dict.fromkeys(users_old))
                return self.collection.update(
                    {"_id": chat_id},
                    {
                        "_id": chat_id,
                        "chat_name": chat_name,
                        "users": users,
                    },
                )
            return self.collection.insert_one(
                {
                    "_id": chat_id,
                    "chat_name": chat_name,
                    "users": [user_id],
                },
            )

    def count_chat_users(self, chat_id: int):
        with INSERTION_LOCK:
            curr = self.collection.find_one({"_id": chat_id})
            if curr:
                return len(curr["users"])
            return 0

    def chat_members(self, chat_id: int):
        with INSERTION_LOCK:
            curr = self.collection.find_one({"_id": chat_id})
            if curr:
                return curr["users"]
            return []

    def count_chats(self):
        with INSERTION_LOCK:
            return self.collection.count()

    def list_chats(self):
        with INSERTION_LOCK:
            chats = self.collection.find_all()
            chat_list = []
            for chat in chats:
                chat_list.append(chat["_id"])
            return chat_list

    def get_all_chats(self):
        with INSERTION_LOCK:
            return self.collection.find_all()

    # Migrate if chat id changes!
    def migrate_chat(self, old_chat_id: int, new_chat_id: int):
        with INSERTION_LOCK:

            old_chat_db = self.collection.find_one({"_id": old_chat_id})
            if old_chat_db:
                new_data = old_chat_db.update({"_id": new_chat_id})
                self.collection.delete_one({"_id": old_chat_id})
                self.collection.insert_one(new_data)
            return
