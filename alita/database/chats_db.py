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

CHATS_CACHE = {}


class Chats:
    """Class to manage users for bot."""

    def __init__(self) -> None:
        self.collection = MongoDB("chats")

    def remove_chat(self, chat_id: int):
        with INSERTION_LOCK:
            self.collection.delete_one({"_id": chat_id})

    def update_chat(self, chat_id: int, chat_name: str, user_id: int):
        global CHATS_CACHE
        with INSERTION_LOCK:

            # Local Cache
            try:
                chat = CHATS_CACHE[chat_id]
                users_old = chat["users"]
                if user_id in set(users_old):
                    # If user_id already exists, return
                    return "user already exists in chat users"
                users_old.append(user_id)
                users = list(set(users_old))
                CHATS_CACHE[chat_id] = {
                    "chat_name": chat_name,
                    "users": users,
                }
            except KeyError:
                pass

            # Databse Cache
            curr = self.collection.find_one({"_id": chat_id})
            if curr:
                users_old = curr["users"]
                users_old.append(user_id)
                users = list(set(users_old))
                return self.collection.update(
                    {"_id": chat_id},
                    {
                        "_id": chat_id,
                        "chat_name": chat_name,
                        "users": users,
                    },
                )

            CHATS_CACHE[chat_id] = {
                "chat_name": chat_name,
                "users": [user_id],
            }
            return self.collection.insert_one(
                {
                    "_id": chat_id,
                    "chat_name": chat_name,
                    "users": [user_id],
                },
            )

    def count_chat_users(self, chat_id: int):
        with INSERTION_LOCK:
            try:
                return len(CHATS_CACHE[chat_id]["users"])
            except Exception as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())
                curr = self.collection.find_one({"_id": chat_id})
                if curr:
                    return len(curr["users"])
            return 0

    def chat_members(self, chat_id: int):
        with INSERTION_LOCK:
            try:
                return CHATS_CACHE[chat_id]["users"]
            except Exception as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())
                curr = self.collection.find_one({"_id": chat_id})
                if curr:
                    return curr["users"]
            return []

    def count_chats(self):
        with INSERTION_LOCK:
            try:
                return len(CHATS_CACHE)
            except Exception as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())
                return self.collection.count() or 0

    def list_chats(self):
        with INSERTION_LOCK:
            try:
                return list(CHATS_CACHE.keys())
            except Exception as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())
                chats = self.collection.find_all()
                chat_list = {i["_id"] for i in chats}
                return list(chat_list)

    def get_all_chats(self):
        with INSERTION_LOCK:
            try:
                return CHATS_CACHE
            except Exception as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())
                return self.collection.find_all()

    def get_chat_info(self, chat_id: int):
        with INSERTION_LOCK:
            try:
                return CHATS_CACHE[chat_id]
            except Exception as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())
                return self.collection.find_one({"_id": chat_id})

    def load_from_db(self):
        with INSERTION_LOCK:
            return self.collection.find_all()

    # Migrate if chat id changes!
    def migrate_chat(self, old_chat_id: int, new_chat_id: int):
        global CHATS_CACHE
        with INSERTION_LOCK:

            # Update locally
            try:
                old_db_local = CHATS_CACHE[old_chat_id]
                del CHATS_CACHE[old_chat_id]
                CHATS_CACHE[new_chat_id] = old_db_local
            except KeyError:
                pass

            # Update in db
            old_chat_db = self.collection.find_one({"_id": old_chat_id})
            if old_chat_db:
                new_data = old_chat_db.update({"_id": new_chat_id})
                self.collection.delete_one({"_id": old_chat_id})
                self.collection.insert_one(new_data)


def __load_chats_cache():
    global CHATS_CACHE
    start = time()
    db = Chats()
    chats = db.load_from_db()
    CHATS_CACHE = {
        int(chat["_id"]): {
            "chat_name": chat["chat_name"],
            "users": chat["users"],
        }
        for chat in chats
    }
    LOGGER.info(f"Loaded Chats Cache - {round((time()-start),3)}s")
