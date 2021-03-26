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
from time import time
from traceback import format_exc

from alita import LOGGER
from alita.database import MongoDB

INSERTION_LOCK = RLock()


ANTISPAM_BANNED = set()


class GBan:
    """Class for managing Gbans in bot."""

    def __init__(self) -> None:
        self.collection = MongoDB("gbans")

    def check_gban(self, user_id: int):
        with INSERTION_LOCK:
            if user_id in ANTISPAM_BANNED:
                return True
            return bool(self.collection.find_one({"_id": user_id}))

    def add_gban(self, user_id: int, reason: str, by_user: int):
        global ANTISPAM_BANNED
        with INSERTION_LOCK:

            # Check if  user is already gbanned or not
            if self.collection.find_one({"_id": user_id}):
                return self.update_gban_reason(user_id, reason)

            # If not already gbanned, then add to gban
            time_rn = datetime.now()
            ANTISPAM_BANNED.add(user_id)
            return self.collection.insert_one(
                {
                    "_id": user_id,
                    "reason": reason,
                    "by": by_user,
                    "time": time_rn,
                },
            )

    def remove_gban(self, user_id: int):
        global ANTISPAM_BANNED
        with INSERTION_LOCK:
            # Check if  user is already gbanned or not
            if self.collection.find_one({"_id": user_id}):
                ANTISPAM_BANNED.remove(user_id)
                return self.collection.delete_one({"_id": user_id})

            return "User not gbanned!"

    def get_gban(self, user_id: int):
        if self.check_gban(user_id):
            curr = self.collection.find_one({"_id": user_id})
            if curr:
                return True, curr["reason"]
        return False, ""

    def update_gban_reason(self, user_id: int, reason: str):
        with INSERTION_LOCK:
            return self.collection.update(
                {"_id": user_id},
                {"reason": reason},
            )

    def count_gbans(self):
        with INSERTION_LOCK:
            try:
                return len(ANTISPAM_BANNED)
            except Exception as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())
                return self.collection.count()

    def load_from_db(self):
        with INSERTION_LOCK:
            return self.collection.find_all()

    def list_gbans(self):
        with INSERTION_LOCK:
            try:
                return list(ANTISPAM_BANNED)
            except Exception as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())
            return self.collection.find_all()


def __load_antispam_users():
    global ANTISPAM_BANNED
    start = time()
    db = GBan()
    users = db.load_from_db()
    ANTISPAM_BANNED = {i["_id"] for i in users}
    LOGGER.info(f"Loaded AntispamBanned Cache - {round((time()-start),3)}s")
