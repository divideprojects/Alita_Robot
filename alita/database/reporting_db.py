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

REPORTING_CACHE = {}


class Reporting:
    """Class for managing report settings of users and groups."""

    db_name = "reporting"

    def __init__(self, chat_id: int) -> None:
        self.collection = MongoDB(self.db_name)
        self.chat_id = chat_id
        self.chat_info = self.__ensure_in_db()

    def get_chat_type(self):
        if str(self.chat_id).startswith("-100"):
            chat_type = "supergroup"
        else:
            chat_type = "user"
        return chat_type

    def set_settings(self, status: bool = True):
        with INSERTION_LOCK:
            self.chat_info["status"] = status
            return self.collection.update(
                {"_id": self.chat_id},
                {"status": self.chat_info["status"]},
            )

    def get_settings(self):
        with INSERTION_LOCK:
            return self.chat_info["status"]

    @staticmethod
    def load_from_db():
        with INSERTION_LOCK:
            collection = MongoDB(Reporting.db_name)
            return collection.find_all() or []

    def __ensure_in_db(self):
        chat_data = self.collection.find_one({"_id": self.chat_id})
        if not chat_data:
            chat_type = self.get_chat_type()
            new_data = {"_id": self.chat_id, "status": True, "chat_type": chat_type}
            self.collection.insert_one(new_data)
            LOGGER.info(f"Initialized Language Document for chat {self.chat_id}")
            return new_data
        return chat_data

    # Migrate if chat id changes!
    def migrate_chat(self, new_chat_id: int):
        old_chat_db = self.collection.find_one({"_id": self.chat_id})
        new_data = old_chat_db.update({"_id": new_chat_id})
        self.collection.insert_one(new_data)
        self.collection.delete_one({"_id": self.chat_id})

    @staticmethod
    def repair_db(collection):
        all_data = collection.find_all()
        keys = {"status": True, "chat_type": ""}
        for data in all_data:
            for key, val in keys.items():
                try:
                    _ = data[key]
                except KeyError:
                    LOGGER.warning(
                        f"Repairing Reporting Database - setting '{key}:{val}' for {data['_id']}",
                    )
                    collection.update({"_id": data["_id"]}, {key: val})


def __check_db_status():
    LOGGER.info("Starting Reports Database Repair...")
    collection = MongoDB(Reporting.db_name)
    Reporting.repair_db(collection)


__check_db_status()
