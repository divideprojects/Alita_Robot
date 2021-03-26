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
from alita.utils.caching import ADMIN_CACHE

INSERTION_LOCK = RLock()

APPROVE_CACHE = {}


class Approve:
    """Class for managing Approves in Chats in Bot."""

    def __init__(self) -> None:
        self.collection = MongoDB("approve")

    def check_approve(self, chat_id: int, user_id: int):
        with INSERTION_LOCK:

            try:
                users = list(APPROVE_CACHE[chat_id])
            except KeyError:
                return True

            if user_id in {i[0] for i in users}:
                return True

            curr_approve = self.collection.find_one(
                {"_id": chat_id},
            )
            if curr_approve:
                try:
                    return next(
                        user for user in curr_approve["users"] if user[0] == user_id
                    )
                except StopIteration:
                    return False

            return False

    def add_approve(self, chat_id: int, user_id: int, user_name: str):
        global APPROVE_CACHE
        with INSERTION_LOCK:

            try:
                users = list(APPROVE_CACHE[chat_id])
            except KeyError:
                return True

            if user_id in {i[0] for i in users}:
                return True

            users_old = APPROVE_CACHE[chat_id]
            users_old.add((user_id, user_name))
            APPROVE_CACHE[chat_id] = users_old

            curr = self.collection.find_one({"_id": chat_id})
            if curr:
                users_old = curr["users"]
                users_old.append((user_id, user_name))
                users = list(set(users_old))  # Remove duplicates
                return self.collection.update(
                    {"_id": chat_id},
                    {
                        "_id": chat_id,
                        "users": users,
                    },
                )

            APPROVE_CACHE[chat_id] = {user_id, user_name}
            return self.collection.insert_one(
                {
                    "_id": chat_id,
                    "users": [(user_id, user_name)],
                },
            )

    def remove_approve(self, chat_id: int, user_id: int):
        global APPROVE_CACHE
        with INSERTION_LOCK:

            try:
                users = list(APPROVE_CACHE[chat_id])
            except KeyError:
                return True
            try:
                user = next(user for user in users if user[0] == user_id)
                users.remove(user)
                ADMIN_CACHE[chat_id] = users
            except StopIteration:
                pass

            curr = self.collection.find_one({"_id": chat_id})
            if curr:
                users = curr["users"]
                try:
                    user = next(user for user in users if user[0] == user_id)
                except StopIteration:
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
        global APPROVE_CACHE
        with INSERTION_LOCK:
            del APPROVE_CACHE[chat_id]
            return self.collection.delete_one(
                {"_id": chat_id},
            )

    def list_approved(self, chat_id: int):
        with INSERTION_LOCK:
            try:
                return APPROVE_CACHE[chat_id]
            except KeyError:
                pass
            except Exception as ef:
                curr = self.collection.find_one({"_id": chat_id})
                if curr:
                    return curr["users"]
                LOGGER.error(ef)
                LOGGER.error(format_exc())
            return []

    def count_all_approved(self):
        with INSERTION_LOCK:
            try:
                return len(set(list(ADMIN_CACHE.keys())))
            except KeyError:
                pass
            except Exception as ef:
                num = 0
                curr = self.collection.find_all()
                if curr:
                    for chat in curr:
                        users = chat["users"]
                        num += len(users)
                LOGGER.error(ef)
                LOGGER.error(format_exc())

            return num

    def count_approved_chats(self):
        with INSERTION_LOCK:
            try:
                return len(list(APPROVE_CACHE.keys()))
            except Exception as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())
                return (self.collection.count()) or 0

    def count_approved(self, chat_id: int):
        with INSERTION_LOCK:
            try:
                return len(APPROVE_CACHE[chat_id])
            except KeyError:
                pass
            except Exception as ef:
                all_app = self.collection.find_one({"_id": chat_id})
                if all_app:
                    return len(all_app["users"]) or 0
                LOGGER.error(ef)
                LOGGER.error(format_exc())
            return 0

    def load_from_db(self):
        return self.collection.find_all()

    # Migrate if chat id changes!
    def migrate_chat(self, old_chat_id: int, new_chat_id: int):
        global APPROVE_CACHE
        with INSERTION_LOCK:

            # Update locally
            try:
                old_db_local = APPROVE_CACHE[old_chat_id]
                del APPROVE_CACHE[old_chat_id]
                APPROVE_CACHE[new_chat_id] = old_db_local
            except KeyError:
                pass

            old_chat_db = self.collection.find_one({"_id": old_chat_id})
            if old_chat_db:
                new_data = old_chat_db.update({"_id": new_chat_id})
                self.collection.delete_one({"_id": old_chat_id})
                self.collection.insert_one(new_data)


def __load_approve_cache():
    global APPROVE_CACHE
    start = time()
    db = Approve()
    all_approved = db.load_from_db()

    APPROVE_CACHE = {chat["_id"]: chat["users"] for chat in all_approved}
    LOGGER.info(f"Loaded Approve Cache - {round((time()-start),3)}s")
