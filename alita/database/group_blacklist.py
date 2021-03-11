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
from alita.database.chats_db import Chats

INSERTION_LOCK = RLock()

BLACKLIST_CHATS = []


class GroupBlacklist:
    """Class to blacklist chats where bot will exit."""

    def __init__(self) -> None:
        self.collection = MongoDB("group_blacklists")

    def add_chat(self, chat_id: int):
        with INSERTION_LOCK:
            global BLACKLIST_CHATS
            Chats().remove_chat(chat_id)  # Delete chat from database
            BLACKLIST_CHATS.append(chat_id)
            BLACKLIST_CHATS.sort()
            return self.collection.insert_one({"_id": chat_id})

    def remove_chat(self, chat_id: int):
        with INSERTION_LOCK:
            global BLACKLIST_CHATS
            BLACKLIST_CHATS.remove(chat_id)
            BLACKLIST_CHATS.sort()
            return self.collection.delete_one({"_id": chat_id})

    def list_all_chats(self):
        with INSERTION_LOCK:
            try:
                BLACKLIST_CHATS.sort()
                return BLACKLIST_CHATS
            except Exception:
                return self.collection.find_all()
