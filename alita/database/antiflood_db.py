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


class AntiFlood:
    """Class for managing antiflood in groups."""

    def __init__(self) -> None:
        self.collection = MongoDB("antiflood")

    def get_grp(self, chat_id: int):
        with INSERTION_LOCK:
            return self.collection.find_one({"_id": chat_id})

    def set_status(self, chat_id: int, status: bool = False):
        with INSERTION_LOCK:

            chat_dict = self.get_grp(chat_id)
            if chat_dict:
                return self.collection.update(
                    {"_id": chat_id},
                    {"status": status},
                )

            return self.collection.insert_one({"_id": chat_id, "status": status})

    def get_status(self, chat_id: int):
        with INSERTION_LOCK:
            z = self.get_grp(chat_id)
            if z:
                return z["status"]
            return False

    def set_antiflood(self, chat_id: int, max_msg: int):
        with INSERTION_LOCK:

            chat_dict = self.get_grp(chat_id)
            if chat_dict:
                return self.collection.update(
                    {"_id": chat_id},
                    {"max_msg": max_msg},
                )

            return self.collection.insert_one({"_id": chat_id, "max_msg": max_msg})

    def get_antiflood(self, chat_id: int):
        with INSERTION_LOCK:
            z = self.get_grp(chat_id)
            if z:
                return z["max_msg"]
            return 0

    def set_action(self, chat_id: int, action: str = "mute"):
        with INSERTION_LOCK:

            if action not in ("kick", "ban", "mute"):
                action = "mute"  # Default action

            chat_dict = self.get_grp(chat_id)
            if chat_dict:

                return self.collection.update(
                    {"_id": chat_id},
                    {"action": action},
                )

            return self.collection.insert_one({"_id": chat_id, "action": action})

    def get_action(self, chat_id: int):
        with INSERTION_LOCK:

            z = self.get_grp(chat_id)
            if z:
                return z["action"]
        return "none"

    def get_all_antiflood_settings(self):
        return self.collection.find_all()

    def get_num_antiflood(self):
        return self.collection.count()

    # Migrate if chat id changes!
    def migrate_chat(self, old_chat_id: int, new_chat_id: int):
        with INSERTION_LOCK:

            old_chat_db = self.collection.find_one({"_id": old_chat_id})
            if old_chat_db:
                new_data = old_chat_db.update({"_id": new_chat_id})
                self.collection.delete_one({"_id": old_chat_id})
                self.collection.insert_one(new_data)
