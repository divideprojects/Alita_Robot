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
from traceback import format_exc

from alita import LOGGER
from alita.database import MongoDB

INSERTION_LOCK = RLock()

ANTIPIN_CHATS = []


class AntiChannelPin:
    """Class for managing antichannelpins in chats."""

    def __init__(self) -> None:
        self.collection = MongoDB("antichannelpin")

    def check_antipin(self, chat_id: int):
        with INSERTION_LOCK:

            if chat_id in ANTIPIN_CHATS:
                return True

            curr = self.collection.find_one({"_id": chat_id})
            if curr:
                return True

            return False

    def toggle_antipin(self, chat_id: int, status: bool = False):
        global ANTIPIN_CHATS
        with INSERTION_LOCK:
            if self.check_antipin(chat_id):
                if status:
                    if not chat_id in ANTIPIN_CHATS:
                        ANTIPIN_CHATS.append(chat_id)
                else:
                    ANTIPIN_CHATS.remove(chat_id)
                return self.collection.update({"_id": chat_id}, {"status": status})
            ANTIPIN_CHATS.append(chat_id)
            return self.collection.insert_one({"_id": chat_id, "status": status})

    def count_antipin_chats(self):
        with INSERTION_LOCK:
            try:
                return len(ANTIPIN_CHATS)
            except Exception as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())
                return self.collection.count({"status": True})

    def list_antipin_chats(self):
        with INSERTION_LOCK:
            try:
                return ANTIPIN_CHATS
            except Exception as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())
                return self.collection.find_all({"status": True})


def __load_antichannelpin_chats():
    global ANTIPIN_CHATS
    db = AntiChannelPin()
    antipin_chats = db.list_antipin_chats()
    for chat in antipin_chats:
        ANTIPIN_CHATS.append(chat["_id"])


__load_antichannelpin_chats()
