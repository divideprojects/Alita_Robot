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


class SpamProtect:
    """Class for managing Spam Protection settings of chats!"""

    def __init__(self) -> None:
        self.collection = MongoDB("spam_protect")

    def get_cas_status(self, chat_id: int):
        with INSERTION_LOCK:
            curr = self.collection.find_one({"_id": chat_id})
            if curr:
                stat = curr["cas"]
                return stat
            self.collection.insert_one(
                {"_id": chat_id, "cas": False, "underattack": False},
            )
            return False

    def set_cas_status(self, chat_id: int, status: bool = False):
        with INSERTION_LOCK:
            curr = self.collection.find_one({"_id": chat_id})
            if curr:
                return self.collection.update(
                    {"_id": chat_id},
                    {"_id": chat_id, "cas": status},
                )
            self.collection.insert_one(
                {"_id": chat_id, "cas": status, "underattack": False},
            )
            return status

    def get_attack_status(self, chat_id: int):
        with INSERTION_LOCK:
            curr = self.collection.find_one({"_id": chat_id})
            if curr:
                stat = curr["underattack"]
                return stat
            self.collection.insert_one(
                {"_id": chat_id, "cas": False, "underattack": False},
            )
            return False

    def set_attack_status(self, chat_id: int, status: bool = False):
        with INSERTION_LOCK:
            curr = self.collection.find_one({"_id": chat_id})
            if curr:
                return self.collection.update(
                    {"_id": chat_id},
                    {"_id": chat_id, "underattack": status},
                )
            self.collection.insert_one(
                {"_id": chat_id, "cas": False, "underattack": status},
            )
            return status

    def get_cas_enabled_chats_num(self):
        with INSERTION_LOCK:
            curr = self.collection.find_all()
            num = 0
            if curr:
                for chat in curr:
                    if chat["cas"]:
                        num += 1
            return num

    def get_attack_enabled_chats_num(self):
        with INSERTION_LOCK:
            curr = self.collection.find_all()
            num = 0
            if curr:
                for chat in curr:
                    if chat["underattack"]:
                        num += 1
            return num

    def get_cas_enabled_chats(self):
        with INSERTION_LOCK:
            curr = self.collection.find_all()
            lst = []
            if curr:
                for chat in curr:
                    if chat["cas"]:
                        lst.append(chat["_id"])
            return lst

    def get_attack_enabled_chats(self):
        with INSERTION_LOCK:
            curr = self.collection.find_all()
            lst = []
            if curr:
                for chat in curr:
                    if chat["underattack"]:
                        lst.append(chat["_id"])
            return lst

    # Migrate if chat id changes!
    def migrate_chat(self, old_chat_id: int, new_chat_id: int):
        with INSERTION_LOCK:

            old_chat_db = self.collection.find_one({"_id": old_chat_id})
            if old_chat_db:
                new_data = old_chat_db.update({"_id": new_chat_id})
                self.collection.delete_one({"_id": old_chat_id})
                self.collection.insert_one(new_data)
            return
