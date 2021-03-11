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


class Users:
    """Class to manage users for bot."""

    def __init__(self) -> None:
        self.collection = MongoDB("users")

    def update_user(self, user_id: int, name: str, username: str = None):
        with INSERTION_LOCK:
            curr = self.collection.find_one({"_id": user_id})
            if curr:
                return self.collection.update(
                    {"_id": user_id},
                    {"username": username, "name": name},
                )
            return self.collection.insert_one(
                {"_id": user_id, "username": username, "name": name},
            )

    def delete_user(self, user_id: int):
        with INSERTION_LOCK:
            curr = self.collection.find_one({"_id": user_id})
            if curr:
                return self.collection.delete_one(
                    {"_id": user_id},
                )
            return True

    def count_users(self):
        with INSERTION_LOCK:
            return self.collection.count()

    def list_users(self):
        with INSERTION_LOCK:
            return self.collection.find_all()

    def get_user_info(self, user_id):
        with INSERTION_LOCK:
            if isinstance(user_id, int) and not user_id.startswith("@"):
                curr = self.collection.find_one({"_id": user_id})
            elif isinstance(user_id, str) and user_id.startswith("@"):
                user_id = user_id.replace("@", "")
                curr = self.collection.find_one({"username": user_id})
            else:
                curr = None
            if curr:
                return curr
            return {}
