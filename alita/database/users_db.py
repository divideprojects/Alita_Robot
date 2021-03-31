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


class Users:
    """Class to manage users for bot."""

    db_name = "users"

    def __init__(self, user_id: int) -> None:
        self.collection = MongoDB(self.db_name)
        self.user_id = user_id
        self.user_info = self.__ensure_in_db()

    def update_user(self, name: str, username: str = None):
        with INSERTION_LOCK:
            if (
                name == self.user_info["name"]
                and username == self.user_info["username"]
            ):
                return True
            return self.collection.update(
                {"_id": self.user_id},
                {"username": username, "name": name},
            )

    def delete_user(self):
        with INSERTION_LOCK:
            return self.collection.delete_one(
                {"_id": self.user_id},
            )

    @staticmethod
    def count_users():
        with INSERTION_LOCK:
            collection = MongoDB(Users.db_name)
            return collection.count()

    def get_my_info(self):
        with INSERTION_LOCK:
            return self.user_id

    @staticmethod
    def list_users():
        with INSERTION_LOCK:
            collection = MongoDB(Users.db_name)
            return collection.find_all()

    @staticmethod
    def get_user_info(user_id: int or str):
        with INSERTION_LOCK:
            collection = MongoDB(Users.db_name)
            if isinstance(user_id, int):
                curr = collection.find_one({"_id": user_id})
            elif isinstance(user_id, str):
                # user_id[1:] because we don't want the '@' in the username search!
                curr = collection.find_one({"username": user_id[1:]})
            else:
                curr = None

            if curr:
                return curr

            return {}

    @staticmethod
    def load_from_db():
        with INSERTION_LOCK:
            collection = MongoDB(Users.db_name)
            return collection.find_all()

    def __ensure_in_db(self):
        chat_data = self.collection.find_one({"_id": self.chat_id})
        if not chat_data:
            chat_type = self.get_chat_type(self.chat_id)
            new_data = {"_id": self.chat_id, "status": True, "chat_type": chat_type}
            self.collection.insert_one(new_data)
            LOGGER.info(f"Initialized Language Document for chat {self.chat_id}")
            return new_data
        return chat_data
