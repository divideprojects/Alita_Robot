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


class AFK:
    """Class for managing AFKs of users."""

    def __init__(self) -> None:
        self.collection = MongoDB("afk_users")

    async def check_afk(self, user_id: int):
        return await self.collection.find_one({"user_id": user_id})

    async def add_afk(self, user_id: int, reason: str = None):
        if await self.check_afk(user_id):
            return await self.collection.update(
                {"user_id": user_id},
                {"user_id": user_id, "reason": reason},
            )
        return await self.collection.insert_one({"user_id": user_id, "reason": reason})

    async def remove_afk(self, user_id: int):
        if await self.check_afk(user_id):
            return await self.collection.delete_one({"user_id": user_id})
        return

    async def count_afk(self):
        return await self.collection.count()

    async def list_afk_users(self):
        return await self.collection.find_all()
