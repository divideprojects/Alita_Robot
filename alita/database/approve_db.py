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


class Approve:
    """Class for managing Approves in Chats in Bot."""

    def __init__(self) -> None:
        self.collection = MongoDB("approve")

    def check_approve(self, chat_id: int, user_id: int):
        with INSERTION_LOCK:
            curr_approve = self.collection.find_one(
                {"_id": chat_id},
            )
            if curr_approve:
                try:
                    return next(
                        user for user in curr_approve["users"] if user[0] == user_id
                    )
                except Exception:
                    return False

            return False

    def add_approve(self, chat_id: int, user_id: int, user_name: str):
        with INSERTION_LOCK:
            curr = self.collection.find_one({"_id": chat_id})
            if curr:
                users_old = curr["users"]
                users_old.append((user_id, user_name))
                users = list(dict.fromkeys(users_old))  # Remove duplicates
                return self.collection.update(
                    {"_id": chat_id},
                    {
                        "_id": chat_id,
                        "users": users,
                    },
                )
            return self.collection.insert_one(
                {
                    "_id": chat_id,
                    "users": [(user_id, user_name)],
                },
            )

    def remove_approve(self, chat_id: int, user_id: int):
        with INSERTION_LOCK:
            curr = self.collection.find_one({"_id": chat_id})
            if curr:
                users = curr["users"]

                try:
                    user = next(user for user in users if user[0] == user_id)
                except Exception:
                    return "Not Approved"

                users.remove(user)

                # If the list is emptied, then delete it
                if not users:
                    return self.collection.delete_one(
                        {"_id": chat_id},
                    )

                return self.collection.update(
                    {"_id": chat_id},
                    {
                        "_id": chat_id,
                        "users": users,
                    },
                )
            return "Not approved"

    def unapprove_all(self, chat_id: int):
        with INSERTION_LOCK:
            return self.collection.delete_one(
                {"_id": chat_id},
            )

    def list_approved(self, chat_id: int):
        with INSERTION_LOCK:
            if self.collection.find_one({"_id": chat_id}):
                return (self.collection.find_one({"_id": chat_id}))["users"]
            return []

    def count_all_approved(self):
        with INSERTION_LOCK:
            num = 0
            curr = self.collection.find_all()
            if curr:
                for chat in curr:
                    users = chat["users"]
                    num += len(users)

            return num

    def count_approved_chats(self):
        with INSERTION_LOCK:
            return (self.collection.count()) or 0

    def count_approved(self, chat_id: int):
        with INSERTION_LOCK:
            all_app = self.collection.find_one({"_id": chat_id})
            return len(all_app["users"]) or 0

    # Migrate if chat id changes!
    def migrate_chat(self, old_chat_id: int, new_chat_id: int):
        with INSERTION_LOCK:

            old_chat_db = self.collection.find_one({"_id": old_chat_id})
            if old_chat_db:
                new_data = old_chat_db.update({"_id": new_chat_id})
                self.collection.delete_one({"_id": old_chat_id})
                self.collection.insert_one(new_data)
            return
