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

    async def add_blacklist(self, chat_id: int, trigger: str):
        curr = await self.collection.find_one({"chat_id": chat_id})
        if curr:
            triggers_old = curr["triggers"]
            triggers_old.append(trigger)
            triggers = list(dict.fromkeys(triggers_old))
            return await self.collection.update(
                {"chat_id": chat_id},
                {
                    "chat_id": chat_id,
                    "triggers": triggers,
                },
            )
        return await self.collection.insert_one(
            {
                "chat_id": chat_id,
                "triggers": [trigger],
                "action": "mute",
            },
        )

    async def remove_blacklist(self, chat_id: int, trigger: str):
        curr = await self.collection.find_one({"chat_id": chat_id})
        if curr:
            triggers_old = curr["triggers"]
            try:
                triggers_old.remove(trigger)
            except ValueError:
                return False
            triggers = list(dict.fromkeys(triggers_old))
            return await self.collection.update(
                {"chat_id": chat_id},
                {
                    "chat_id": chat_id,
                    "triggers": triggers,
                },
            )

    async def get_blacklists(self, chat_id: int):
        curr = await self.collection.find_one({"chat_id": chat_id})
        if curr:
            return curr["triggers"]
        return []

    async def count_blacklists_all(self):
        curr = await self.collection.find_all()
        num = 0
        for chat in curr:
            num += len(chat["triggers"])
        return num

    async def count_blackists_chats(self):
        curr = await self.collection.find_all()
        num = 0
        for chat in curr:
            if chat["triggers"]:
                num += 1
        return num

    async def set_action(self, chat_id: int, action: int):

        if action not in ("kick", "mute", "ban", "warn"):
            return "invalid action"

        curr = await self.collection.find_one({"chat_id": chat_id})
        if curr:
            return await self.collection.update(
                {"chat_id": chat_id},
                {"chat_id": chat_id, "action": action},
            )
        return await self.collection.insert_one(
            {
                "chat_id": chat_id,
                "triggers": [],
                "action": action,
            },
        )

    async def get_action(self, chat_id: int):
        curr = await self.collection.find_one({"chat_id": chat_id})
        if curr:
            return curr["action"] or "mute"
        await self.collection.insert_one(
            {
                "chat_id": chat_id,
                "triggers": [],
                "action": "mute",
            },
        )
        return "mute"

    # Migrate if chat id changes!
    async def migrate_chat(self, old_chat_id: int, new_chat_id: int):
        old_chat = await self.collection.find_one({"chat_id": old_chat_id})
        if old_chat:
            return await self.collection.update(
                {"chat_id": old_chat_id},
                {"chat_id": new_chat_id},
            )
        return
