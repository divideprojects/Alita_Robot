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


class Pins(MongoDB):
    """Class for managing antichannelpins in chats."""

    # Database name to connect to to preform operations
    db_name = "antichannelpin"

    def __init__(self, chat_id: int) -> None:
        super().__init__(self.db_name)
        self.chat_id = chat_id
        self.chat_info = self.__ensure_in_db()

    def get_settings(self):
        with INSERTION_LOCK:
            return self.chat_info

    def antichannelpin_on(self):
        with INSERTION_LOCK:
            return self.set_on("antichannelpin")

    def cleanlinked_on(self):
        with INSERTION_LOCK:
            return self.set_on("cleanlinked")

    def antichannelpin_off(self):
        with INSERTION_LOCK:
            return self.set_off("antichannelpin")

    def cleanlinked_off(self):
        with INSERTION_LOCK:
            return self.set_off("cleanlinked")

    def set_on(self, atype: str):
        with INSERTION_LOCK:
            otype = "cleanlinked" if atype == "antichannelpin" else "antichannelpin"
            return self.update(
                {"_id": self.chat_id},
                {atype: True, otype: False},
            )

    def set_off(self, atype: str):
        with INSERTION_LOCK:
            otype = "cleanlinked" if atype == "antichannelpin" else "antichannelpin"
            return self.update(
                {"_id": self.chat_id},
                {atype: False, otype: False},
            )

    def __ensure_in_db(self):
        chat_data = self.find_one({"_id": self.chat_id})
        if not chat_data:
            new_data = {
                "_id": self.chat_id,
                "antichannelpin": False,
                "cleanlinked": False,
            }
            self.insert_one(new_data)
            LOGGER.info(f"Initialized Pins Document for chat {self.chat_id}")
            return new_data
        return chat_data

    # Migrate if chat id changes!
    def migrate_chat(self, new_chat_id: int):
        old_chat_db = self.find_one({"_id": self.chat_id})
        new_data = old_chat_db.update({"_id": new_chat_id})
        self.insert_one(new_data)
        self.delete_one({"_id": self.chat_id})

    # ----- Static Methods -----
    @staticmethod
    def count_chats(atype: str):
        with INSERTION_LOCK:
            collection = MongoDB(Pins.db_name)
            return collection.count({atype: True})

    @staticmethod
    def list_chats(query: str):
        with INSERTION_LOCK:
            collection = MongoDB(Pins.db_name)
            return collection.find_all({query: True})

    @staticmethod
    def load_from_db():
        with INSERTION_LOCK:
            collection = MongoDB(Pins.db_name)
            return collection.findall()

    @staticmethod
    def repair_db(collection):
        all_data = collection.find_all()
        keys = {"antichannelpin": False, "cleanlinked": False}
        for data in all_data:
            for key, val in keys.items():
                try:
                    _ = data[key]
                except KeyError:
                    LOGGER.warning(
                        f"Repairing Pins Database - setting '{key}:{val}' for {data['_id']}",
                    )
                    collection.update({"_id": data["_id"]}, {key: val})


def __pre_req_pins_chats():
    LOGGER.info("Starting Pins Database Repair...")
    collection = MongoDB(Pins.db_name)
    Pins.repair_db(collection)
