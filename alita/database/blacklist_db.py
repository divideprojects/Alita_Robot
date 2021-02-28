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


from pyrogram.filters import user
from pyrogram.methods.chats import delete_user_history

from alita.database import MongoDB


class Blacklist:
    """Class to manage database for blacklists for chats."""

    def __init__(self) -> None:
        self.collection = MongoDB("blacklists")

    async def save_blacklist(self, chat_id: int, trigger: str, action: str = "kick"):

        if action not in ("mute", "ban", "kick"):
            action = "kick"

        curr = await self.collection.find_one({"chat_id": chat_id, "trigger": trigger})
        if curr:
            return "Blacklist already added!"
        return await self.collection.insert_one(
            {"chat_id": chat_id, "trigger": trigger, "action": action},
        )

    async def remove_blacklist(self, chat_id: int, trigger: int):
        curr = await self.collection.find_one({"chat_id": chat_id, "trigger": trigger})
        if curr:
            return await self.collection.delete_one(
                {"chat_id": chat_id, "trigger": trigger},
            )
        return "Blacklist not found!"

    async def get_blacklists(self, chat_id: int):
        blacklists = await self.collection.find_all({"chat_id": chat_id})
        return blacklists

    async def set_action(self, chat_id: int, action: str = "kick"):

        if action not in ("mute", "ban", "kick"):
            action = "kick"

        curr_action = (await self.collection.find_all({"actin": action}))["action"]

        if curr_action != action:
            curr = await self.collection.find_all({"chat_id": chat_id})
            if curr:
                for i in curr:
                    await self.collection.update(i, {"action": action})
                return "Updated Action!"
        return f"Current action remains same: {action}"

    async def count_blacklists(self):
        return await self.collection.count()

    async def count_chats_blacklist(self):
        chats_num = await self.collection.find_all()
        chats_ids = []
        for chat in chats_num:
            chats_ids.append(chat["chat_id"])
        return len(list(dict.fromkeys(chats_ids)))

    # Migrate if chat id changes!
    async def migrate_chat(self, old_chat_id: int, new_chat_id: int):
        old_chat = await self.collection.find_one({"chat_id": old_chat_id})
        if old_chat:
            return await self.collection.update(
                {"chat_id": old_chat_id},
                {"chat_id": new_chat_id},
            )
        return
