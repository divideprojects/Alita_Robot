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

    async def set_status(self, chat_id: int, status: bool = False):
        z = (await self.collection.find_one({"chat_id": chat_id}))["status"]
        if z:
            return await self.collection.replace(
                {"chat_id": chat_id},
                {"chat_id": chat_id, "status": status},
            )
        return await self.collection.insert_one({"chat_id": chat_id, "status": status})

    async def antiflood_status(self, chat_id: int):
        return (await self.collection.find_one({"chat_id": chat_id}))["status"]

    async def antiflood_action(self, chat_id: int, action: str = None):
        if action not in ("kick", "ban", "mute", None):
            action = "kick"  # Default action
        z = (await self.collection.find_one({"chat_id": chat_id}))["action"]
        if z:
            return await self.collection.replace(
                {"chat_id": chat_id},
                {"chat_id": chat_id, "action": action},
            )
        return await self.collection.insert_one({"chat_id": chat_id, "action": action})

    async def set_warning(self, chat_id: int, max_msg: int):
        z = (await self.collection.find_one({"chat_id": chat_id}))["max_msg"]
        if z:
            return await self.collection.replace(
                {"chat_id": chat_id},
                {"chat_id": chat_id, "max_msg": max_msg},
            )
        return await self.collection.insert_one(
            {"chat_id": chat_id, "max_msg": max_msg},
        )
