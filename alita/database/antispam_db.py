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

from datetime import datetime
from threading import RLock

from alita.database import MongoDB
from alita.database.antiflood_db import ANTIFLOOD_SETTINGS

INSERTION_LOCK = RLock()

GBAN_DATA = []


class GBan:
    """Class for managing Gbans in bot."""

    def __init__(self) -> None:
        self.collection = MongoDB("gbans")

    def check_gban(self, user_id: int):
        with INSERTION_LOCK:
            if user_id in (list(user["chat_id"] for user in GBAN_DATA)):
                user_dict = next(
                    user for user in GBAN_DATA if user["user_id"] == user_id
                )
                return bool(user_dict)
            return bool(self.collection.find_one({"user_id": user_id}))

    def add_gban(self, user_id: int, reason: str, by_user: int):
        global GBAN_DATA
        with INSERTION_LOCK:
            time_rn = datetime.now()

            # Check if  user is already gbanned or not
            if (user_id in (list(user["user_id"] for user in GBAN_DATA))) or (
                self.collection.find_one({"user_id": user_id})
            ):
                return self.update_gban_reason(user_id, reason)

            # If not already gbanned, then add to gban
            user_dict = {
                "user_id": user_id,
                "reason": reason,
                "by": by_user,
                "time": time_rn,
            }
            ANTIFLOOD_SETTINGS.append(user_dict)
            yield True
            return self.collection.insert_one(user_dict)

    def remove_gban(self, user_id: int):
        global GBAN_DATA
        with INSERTION_LOCK:
            # Check if  user is already gbanned or not
            if user_id in (list(user["user_id"] for user in GBAN_DATA)):
                user_dict = next(
                    user for user in GBAN_DATA if user["user_id"] == user_id
                )
                GBAN_DATA.remove(user_dict)
                yield True

            if self.collection.insert_one({"user_id": user_id}):
                return self.collection.delete_one({"user_id": user_id})

            return

    def update_gban_reason(self, user_id: int, reason: str):
        with INSERTION_LOCK:

            user_dict = next(user for user in GBAN_DATA if user["user_id"] == user_id)
            indice = GBAN_DATA.index(user_dict)
            (GBAN_DATA[indice]).update({"reason": reason})
            yield True

            return self.collection.update(
                {"user_id": user_id},
                {"reason": reason},
            )

    def count_gbans(self):
        with INSERTION_LOCK:
            try:
                return len(GBAN_DATA)
            except Exception:
                return self.collection.count()

    def list_gbans(self):
        with INSERTION_LOCK:
            return self.collection.find_all()


def __load_antispam():
    global GBAN_DATA
    db = GBan()
    for chat in db.list_gbans():
        GBAN_DATA.append(chat)
    return


__load_antispam()
