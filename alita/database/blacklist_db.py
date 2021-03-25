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


class Blacklist:
    """Class to manage database for blacklists for chats."""

    def __init__(self) -> None:
        self.collection = MongoDB("blacklists")

    def add_blacklist(self, chat_id: int, trigger: str):
        with INSERTION_LOCK:
            curr = self.collection.find_one({"_id": chat_id})
            if curr:
                triggers_old = curr["triggers"]
                triggers_old.append(trigger)
                triggers = list(set(triggers_old))
                return self.collection.update(
                    {"_id": chat_id},
                    {
                        "_id": chat_id,
                        "triggers": triggers,
                    },
                )
            return self.collection.insert_one(
                {
                    "_id": chat_id,
                    "triggers": [trigger],
                    "action": "none",
                    "reason": "Automated blacklisted word",
                },
            )

    def remove_blacklist(self, chat_id: int, trigger: str):
        with INSERTION_LOCK:
            curr = self.collection.find_one({"_id": chat_id})
            if curr:
                triggers_old = curr["triggers"]
                try:
                    triggers_old.remove(trigger)
                except ValueError:
                    return "Trigger not found"
                triggers = list(set(triggers_old))
                return self.collection.update(
                    {"_id": chat_id},
                    {
                        "_id": chat_id,
                        "triggers": triggers,
                    },
                )

    def get_blacklists(self, chat_id: int):
        with INSERTION_LOCK:
            curr = self.collection.find_one({"_id": chat_id})
            if curr:
                return curr["triggers"]
            return []

    def count_blacklists_all(self):
        with INSERTION_LOCK:
            curr = self.collection.find_all()
            num = 0
            for chat in curr:
                num += len(chat["triggers"])
            return num

    def count_blackists_chats(self):
        with INSERTION_LOCK:
            curr = self.collection.find_all()
            num = 0
            for chat in curr:
                if chat["triggers"]:
                    num += 1
            return num

    def set_action(self, chat_id: int, action: str):
        with INSERTION_LOCK:

            if action not in ("kick", "mute", "ban", "warn", "none"):
                return "invalid action"

            curr = self.collection.find_one({"_id": chat_id})
            if curr:
                return self.collection.update(
                    {"_id": chat_id},
                    {"_id": chat_id, "action": action},
                )
            return self.collection.insert_one(
                {
                    "_id": chat_id,
                    "triggers": [],
                    "action": action,
                    "reason": "Automated blacklisted word",
                },
            )

    def get_action(self, chat_id: int):
        with INSERTION_LOCK:
            curr = self.collection.find_one({"_id": chat_id})
            if curr:
                return curr["action"] or "none"
            self.collection.insert_one(
                {
                    "_id": chat_id,
                    "triggers": [],
                    "action": "none",
                    "reason": "Automated blacklisted word",
                },
            )
            return "Automated blacklisted word"

    def set_reason(self, chat_id: int, reason: str):
        with INSERTION_LOCK:

            curr = self.collection.find_one({"_id": chat_id})
            if curr:
                return self.collection.update(
                    {"_id": chat_id},
                    {"_id": chat_id, "reason": reason},
                )
            return self.collection.insert_one(
                {
                    "_id": chat_id,
                    "triggers": [],
                    "action": "none",
                    "reason": reason,
                },
            )

    def get_reason(self, chat_id: int):
        with INSERTION_LOCK:
            curr = self.collection.find_one({"_id": chat_id})
            if curr:
                return curr["reason"] or "none"
            self.collection.insert_one(
                {
                    "_id": chat_id,
                    "triggers": [],
                    "action": "none",
                    "reason": "Automated blacklistwd word",
                },
            )
            return "Automated blacklisted word"

    def count_action_bl_all(self, action: str):
        return self.collection.count({"action": action})

    def rm_all_blacklist(self, chat_id: int):
        with INSERTION_LOCK:
            curr = self.collection.find_one({"_id": chat_id})
            if curr:
                self.collection.update(
                    {"_id": chat_id},
                    {"triggers": []},
                )
            return False

    # Migrate if chat id changes!
    def migrate_chat(self, old_chat_id: int, new_chat_id: int):
        with INSERTION_LOCK:

            old_chat_db = self.collection.find_one({"_id": old_chat_id})
            if old_chat_db:
                new_data = old_chat_db.update({"_id": new_chat_id})
                self.collection.delete_one({"_id": old_chat_id})
                self.collection.insert_one(new_data)
