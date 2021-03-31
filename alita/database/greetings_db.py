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


class Greetings:
    """Class for managing antichannelpins in chats."""

    # Database name to connect to to preform operations
    db_name = "welcome_chats"

    def __init__(self, chat_id: int) -> None:
        self.collection = MongoDB(self.db_name)
        self.chat_id = chat_id
        self.chat_info = self.__ensure_in_db()

    # Get settings from database
    def get_welcome_status(self):
        with INSERTION_LOCK:
            return self.chat_info["welcome"]

    def get_goodbye_status(self):
        with INSERTION_LOCK:
            return self.chat_info["goodbye"]

    def get_current_cleanservice_settings(self):
        with INSERTION_LOCK:
            return self.chat_info["cleanservice"]

    def get_current_cleanwelcome_settings(self):
        with INSERTION_LOCK:
            return self.chat_info["cleanwelcome"]

    def get_welcome_text(self):
        with INSERTION_LOCK:
            return self.chat_info["welcome_text"]

    def get_goodbye_text(self):
        with INSERTION_LOCK:
            return self.chat_info["goodbye_text"]

    # Set settings in database
    def set_current_welcome_settings(self, status: bool):
        with INSERTION_LOCK:
            return self.collection.update({"_id": self.chat_id}, {"welcome": status})

    def set_current_goodbye_settings(self, status: bool):
        with INSERTION_LOCK:
            return self.collection.update({"_id": self.chat_id}, {"goodbye": status})

    def set_welcome_text(self, welcome_text: str):
        with INSERTION_LOCK:
            return self.collection.update(
                {"_id": self.chat_id},
                {"welcome_text": welcome_text},
            )

    def set_goodbye_text(self, goodbye_text: str):
        with INSERTION_LOCK:
            return self.collection.update(
                {"_id": self.chat_id},
                {"goodbye_text": goodbye_text},
            )

    def set_current_cleanservice_settings(self, status: bool):
        with INSERTION_LOCK:
            return self.collection.update(
                {"_id": self.chat_id},
                {"cleanservice": status},
            )

    def set_current_cleanwelcome_settings(self, status: bool):
        with INSERTION_LOCK:
            return self.collection.update(
                {"_id": self.chat_id},
                {"cleanwelcome": status},
            )

    def __ensure_in_db(self):
        chat_data = self.collection.find_one({"_id": self.chat_id})
        if not chat_data:
            new_data = {
                "_id": self.chat_id,
                "cleanwelcome": False,
                "cleanservice": False,
                "goodbye_text": "Sad to see you leave {first}.\nTake Care!",
                "welcome_text": "Hey {first}, welcome to {group}!",
                "welcome": True,
                "goodbye": True,
            }
            self.collection.insert_one(new_data)
            LOGGER.info(f"Initialized Greetings Document for chat {self.chat_id}")
            return new_data
        return chat_data

    @staticmethod
    def repair_db(collection):
        all_data = collection.find_all()
        keys = {
            "cleanwelcome": False,
            "cleanservice": False,
            "goodbye_text": "Sad to see you leave {first}.\nTake Care!",
            "welcome_text": "Hey {first}, welcome to {group}!",
            "welcome": True,
            "goodbye": True,
        }
        for data in all_data:
            for key, val in keys.items():
                try:
                    _ = data[key]
                except KeyError:
                    LOGGER.warning(
                        f"Repairing Greeting Database - setting '{key}:{val}' for {data['_id']}",
                    )
                    collection.update({"_id": data["_id"]}, {key: val})


def __check_db_status():
    LOGGER.info("Starting Geeetings Database Repair...")
    collection = MongoDB(Greetings.db_name)
    Greetings.repair_db(collection)


__check_db_status()
