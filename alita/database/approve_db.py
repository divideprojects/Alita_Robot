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


class Approve:
    """Class for managing Approves in Chats in Bot."""

    def __init__(self) -> None:
        self.collection = MongoDB("approve")

    async def check_approve(self, chat_id: int, user_id: int):
        curr_approve = await self.collection.find_one(
            {"chat_id": chat_id, "user_id": user_id},
        )
        if curr_approve:
            return True
        return False

    async def add_approve(self, chat_id: int, user_id: int):
        if await self.collection.find_one({"chat_id": chat_id, "user_id": user_id}):
            return "Already Added!"
        return await self.collection.insert_one(
            {"chat_id": chat_id, "user_id": user_id},
        )

    async def remove_approve(self, chat_id: int, user_id: int):
        if await self.collection.find_one({"chat_id": chat_id, "user_id": user_id}):
            return await self.collection.delete_one(
                {"chat_id": chat_id, "user_id": user_id},
            )
        return "Removed from Approves!"

    async def unapprove_all(self, chat_id: int):
        all_users = await self.collection.find_all({"chat_id": chat_id})
        for user_id in all_users:
            await self.collection.delete_one(
                {"chat_id": chat_id, "user_id": user_id},
            )
        return

    async def list_approved(self, chat_id: int):
        return await self.collection.find_all({"chat_id": chat_id})

    async def count_all_approved(self):
        return await self.collection.count()

    async def count_approved(self, chat_id: int):
        return await self.collection.count({"chat_id": chat_id})

    async def list_all_approved(self):
        return await self.collection.find_all({})
