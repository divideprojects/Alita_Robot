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


from alita.database import MongoDB


class AntiFlood:
    """Class for managing antiflood in groups."""

    def __init__(self) -> None:
        self.collection = MongoDB("antiflood")
        
    async def get_grp(self, chat_id:int):
        return await self.collection.find_one({"chat_id": chat_id})

    async def set_status(self, chat_id: int, status: bool = False):
        if (await self.get_grp(chat_id)):
            return await self.collection.update(
                {"chat_id": chat_id},
                {"status": status},
            )
        return await self.collection.insert_one({"chat_id": chat_id, "status": status})
        
    async get_status(self, chat_id: int):
        z = (await self.get_grp(chat_id))
        if z:
            return z['status']
        return

    async def set_antiflood(self, chat_id: int, max_msg: int):
        if (await self.get_grp(chat_id)):
            return await self.collection.update(
                {"chat_id": chat_id},
                {"max_msg": max_msg},
            )
        return await self.collection.insert_one(
            {"chat_id": chat_id, "max_msg": max_msg},
        )

    async def get_antiflood(self, chat_id: int):
        z = (await self.get_grp(chat_id))
        if z:
            return z['max_msg']
        return

    async def set_action(self, chat_id: int, action: str = "kick"):

        if action not in ("kick", "ban", "mute"):
            action = "kick"  # Default action

        if (await self.get_grp(chat_id)):
            return await self.collection.update(
                {"chat_id": chat_id},
                {"action": action},
            )
        return await self.collection.insert_one({"chat_id": chat_id, "action": action})
        
    async def get_action(self, chat_id: int):
        z = (await self.get_grp(chat_id))
        if z:
            return z['action']
        return
