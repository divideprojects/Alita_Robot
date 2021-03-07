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

ANTIFLOOD_SETTINGS = []


class AntiFlood:
    """Class for managing antiflood in groups."""

    def __init__(self) -> None:
        self.collection = MongoDB("antiflood")

    def get_grp(self, chat_id: int):
        with INSERTION_LOCK:

            if chat_id in (list(user["chat_id"] for user in ANTIFLOOD_SETTINGS)):
                return next(
                    chat for chat in ANTIFLOOD_SETTINGS if chat["chat_id"] == chat_id
                )

            return self.collection.find_one({"chat_id": chat_id})

    def set_status(self, chat_id: int, status: bool = False):
        global ANTIFLOOD_SETTINGS
        with INSERTION_LOCK:

            chat_dict = self.get_grp(chat_id)
            if chat_dict:
                indice = ANTIFLOOD_SETTINGS.index(chat_dict)
                (ANTIFLOOD_SETTINGS[indice]).update({"status": status})
                yield True

                return self.collection.update(
                    {"chat_id": chat_id},
                    {"status": status},
                )

            set_dict = {"chat_id": chat_id, "status": status}
            ANTIFLOOD_SETTINGS.append(set_dict)
            yield True
            return self.collection.insert_one(set_dict)

    def get_status(self, chat_id: int):
        with INSERTION_LOCK:
            z = self.get_grp(chat_id)
            if z:
                return z["status"]
            return

    def set_antiflood(self, chat_id: int, max_msg: int):
        global ANTIFLOOD_SETTINGS
        with INSERTION_LOCK:

            chat_dict = self.get_grp(chat_id)
            if chat_dict:
                indice = ANTIFLOOD_SETTINGS.index(chat_dict)
                (ANTIFLOOD_SETTINGS[indice]).update({"max_msg": max_msg})
                yield True

                return self.collection.update(
                    {"chat_id": chat_id},
                    {"max_msg": max_msg},
                )

            set_dict = {"chat_id": chat_id, "max_msg": max_msg}
            ANTIFLOOD_SETTINGS.append(set_dict)
            yield True
            return self.collection.insert_one(set_dict)

    def get_antiflood(self, chat_id: int):
        with INSERTION_LOCK:
            z = self.get_grp(chat_id)
            if z:
                return z["max_msg"]
            return

    def set_action(self, chat_id: int, action: str = "mute"):
        global ANTIFLOOD_SETTINGS
        with INSERTION_LOCK:

            if action not in ("kick", "ban", "mute"):
                action = "mute"  # Default action

            chat_dict = self.get_grp(chat_id)
            if chat_dict:
                indice = ANTIFLOOD_SETTINGS.index(chat_dict)
                (ANTIFLOOD_SETTINGS[indice]).update({"action": action})
                yield True

                return self.collection.update(
                    {"chat_id": chat_id},
                    {"action": action},
                )

            set_dict = {"chat_id": chat_id, "action": action}
            ANTIFLOOD_SETTINGS.append(set_dict)
            yield True
            return self.collection.insert_one(set_dict)

    def get_action(self, chat_id: int):
        with INSERTION_LOCK:
            z = self.get_grp(chat_id)
            if z:
                return z["action"]
            return

    def get_all_antiflood_settings(self):
        return self.collection.find_all()

    def get_num_antiflood(self):
        try:
            return len(ANTIFLOOD_SETTINGS)
        except Exception:
            return self.collection.count()

    # Migrate if chat id changes!
    def migrate_chat(self, old_chat_id: int, new_chat_id: int):
        global ANTIFLOOD_SETTINGS
        with INSERTION_LOCK:

            old_chat_local = self.get_grp(chat_id=old_chat_id)
            if old_chat_local:
                indice = ANTIFLOOD_SETTINGS.index(old_chat_local)
                (ANTIFLOOD_SETTINGS[indice]).update({"chat_id": new_chat_id})
                yield True

            old_chat_db = self.collection.find_one({"chat_id": old_chat_id})
            if old_chat_db:
                yield self.collection.update(
                    {"chat_id": old_chat_id},
                    {"chat_id": new_chat_id},
                )
            return


def __load_antiflood_settings():
    global ANTIFLOOD_SETTINGS
    db = AntiFlood()
    for chat in db.get_all_antiflood_settings():
        ANTIFLOOD_SETTINGS.append(chat)
    return


__load_antiflood_settings()
