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


class Langs:
    """Class for language options in bot."""

    def __init__(self) -> None:
        self.collection = MongoDB("langs")

    async def get_chat_type(self, chat_id: int):
        if str(chat_id).startswith("-100"):
            chat_type = "supergroup"
        else:
            chat_type = "user"
        return chat_type

    async def set_lang(self, chat_id: int, lang: str = "en"):
        with INSERTION_LOCK:
            chat_type = self.get_chat_type(chat_id)

            if self.collection.find_one({"chat_id": chat_id}):
                return self.collection.update(
                    {"chat_id": chat_id},
                    {"lang": lang},
                )

            return self.collection.insert_one(
                {"chat_id": chat_id, "chat_type": chat_type, "lang": lang},
            )

    async def get_lang(self, chat_id: int):
        with INSERTION_LOCK:
            chat_type = self.get_chat_type(chat_id)

            curr_lang = self.collection.find_one({"chat_id": chat_id})
            if curr_lang:
                return str(curr_lang["lang"])

            self.collection.insert_one(
                {"chat_id": chat_id, "chat_type": chat_type, "lang": "en"},
            )
            return "en"

    # Migrate if chat id changes!
    async def migrate_chat(self, old_chat_id: int, new_chat_id: int):
        with INSERTION_LOCK:
            old_chat = self.collection.find_one({"chat_id": old_chat_id})
            if old_chat:
                return self.collection.update(
                    {"chat_id": old_chat_id},
                    {"chat_id": new_chat_id},
                )
            return
