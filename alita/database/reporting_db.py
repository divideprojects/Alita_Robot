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


class Reporting:
    """Class for managing report settings of users and groups."""

    def __init__(self) -> None:
        self.collection = MongoDB("reporting_settings")

    def get_chat_type(self, chat_id: int):
        _ = self
        if str(chat_id).startswith("-100"):
            chat_type = "supergroup"
        else:
            chat_type = "user"
        return chat_type

    def set_settings(self, chat_id: int, status: bool = True):
        with INSERTION_LOCK:
            chat_type = self.get_chat_type(chat_id)
            curr_settings = (self.collection.find_one({"_id": chat_id}))
            if curr_settings:
                return self.collection.update(
                    {"_id": chat_id},
                    {"status": status},
                )
            return self.collection.insert_one(
                {"_id": chat_id, "chat_type": chat_type, "status": status},
            )

    def get_settings(self, chat_id: int):
        with INSERTION_LOCK:
            chat_type = self.get_chat_type(chat_id)
            curr_settings = (self.collection.find_one({"_id": chat_id}))
            if curr_settings:
                return curr_settings["status"]
            self.collection.insert_one(
                {"_id": chat_id, "chat_type": chat_type, "status": True},
            )
            return True

    # Migrate if chat id changes!
    def migrate_chat(self, old_chat_id: int, new_chat_id: int):
        with INSERTION_LOCK:

            old_chat_db = self.collection.find_one({"_id": old_chat_id})
            if old_chat_db:
                new_data = old_chat_db.update({"_id": new_chat_id})
                self.collection.delete_one({"_id": old_chat_id})
                self.collection.insert_one(new_data)
            return
