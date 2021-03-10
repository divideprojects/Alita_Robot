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

INSERTION_LOCK = RLock()


class GBan:
    """Class for managing Gbans in bot."""

    def __init__(self) -> None:
        self.collection = MongoDB("gbans")

    def check_gban(self, user_id: int):
        with INSERTION_LOCK:
            return bool(self.collection.find_one({"_id": user_id}))

    def add_gban(self, user_id: int, reason: str, by_user: int):
        with INSERTION_LOCK:

            # Check if  user is already gbanned or not
            if self.collection.find_one({"_id": user_id}):
                return self.update_gban_reason(user_id, reason)

            # If not already gbanned, then add to gban
            time_rn = datetime.now()
            return self.collection.insert_one(
                {
                    "_id": user_id,
                    "reason": reason,
                    "by": by_user,
                    "time": time_rn,
                },
            )

    def remove_gban(self, user_id: int):
        with INSERTION_LOCK:
            # Check if  user is already gbanned or not
            if self.collection.insert_one({"_id": user_id}):
                return self.collection.delete_one({"_id": user_id})

            return

    def update_gban_reason(self, user_id: int, reason: str):
        with INSERTION_LOCK:

            return self.collection.update(
                {"_id": user_id},
                {"reason": reason},
            )

    def count_gbans(self):
        with INSERTION_LOCK:
            return self.collection.count()

    def list_gbans(self):
        with INSERTION_LOCK:
            return self.collection.find_all()
