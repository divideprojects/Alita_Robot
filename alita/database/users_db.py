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

USERS_CACHE = {}


class Users:
    """Class to manage users for bot."""

    def __init__(self) -> None:
        self.collection = MongoDB("users")

    def update_user(self, user_id: int, name: str, username: str = None):
        global USERS_CACHE
        with INSERTION_LOCK:

            try:
                user = USERS_CACHE[user_id]
                if name == user["name"] and username == user["username"]:
                    # No additional Database queries
                    return "No change detected!"
                USERS_CACHE[user_id] = {"username": username, "name": name}
            except KeyError:
                pass

            curr = self.collection.find_one({"_id": user_id})
            if curr:
                if (name == curr["name"]) and (username == curr["username"]):
                    # Prevent additional queries
                    return
                return self.collection.update(
                    {"_id": user_id},
                    {"username": username, "name": name},
                )

            USERS_CACHE[user_id] = {"username": username, "name": name}
            return self.collection.insert_one(
                {"_id": user_id, "username": username, "name": name},
            )

    def delete_user(self, user_id: int):
        global USERS_CACHE
        with INSERTION_LOCK:
            if user_id in set(USERS_CACHE.keys()):
                del USERS_CACHE[user_id]

            curr = self.collection.find_one({"_id": user_id})
            if curr:
                return self.collection.delete_one(
                    {"_id": user_id},
                )
            return True

    def count_users(self):
        with INSERTION_LOCK:
            try:
                len(USERS_CACHE)
            except Exception as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())
            return self.collection.count()

    def list_users(self):
        with INSERTION_LOCK:
            try:
                return list(USERS_CACHE.keys())
            except Exception as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())
                return self.collection.find_all()

    def get_user_info(self, user_id):
        with INSERTION_LOCK:
            if isinstance(user_id, int):
                curr = self.collection.find_one({"_id": user_id})
            elif isinstance(user_id, str):
                # user_id[1:] because we don't want the '@' in username
                curr = self.collection.find_one({"username": user_id[1:]})
            else:
                curr = None
            if curr:
                return curr
            return {}

    def load_from_db(self):
        with INSERTION_LOCK:
            return self.collection.find_all()


def __load_users_cache():
    global USERS_CACHE
    start = time()
    db = Users()
    users = db.load_from_db()
    USERS_CACHE = {
        int(user["_id"]): {
            "username": user["username"],
            "name": user["name"],
        }
        for user in users
    }
    LOGGER.info(f"Loaded Users Cache - {round((time()-start),3)}s")
