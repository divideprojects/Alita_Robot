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


class Chats:
    """Class to manage users for bot."""

    def __init__(self) -> None:
        self.collection = MongoDB("chats")

    async def remove_chat(self, chat_id: int):
        await self.collection.delete_one({"chat_id": chat_id})

    async def update_chat(self, chat_id: int, chat_name: str, user_id: int):
        curr = await self.collection.find_one({"chat_id": chat_id})
        if curr:
            users_old = curr["users"]
            users_old.append(user_id)
            users = list(dict.fromkeys(users_old))
            return await self.collection.update(
                {"chat_id": chat_id},
                {
                    "chat_id": chat_id,
                    "chat_name": chat_name,
                    "users": users,
                },
            )
        return await self.collection.insert_one(
            {
                "chat_id": chat_id,
                "chat_name": chat_name,
                "users": [user_id],
            },
        )

    async def count_chat_users(self, chat_id: int):
        curr = await self.collection.find_one({"chat_id": chat_id})
        if curr:
            return len(curr["users"])
        return 0

    async def chat_members(self, chat_id: int):
        curr = await self.collection.find_one({"chat_id": chat_id})
        if curr:
            return curr["users"]
        return []

    async def count_chats(self):
        return await self.collection.count()

    async def list_chats(self):
        chats = await self.collection.find_all()
        chat_list = []
        for chat in chats:
            chat_list.append(chat["chat_id"])
        return chat_list

    # Migrate if chat id changes!
    async def migrate_chat(self, old_chat_id: int, new_chat_id: int):
        old_chat = await self.collection.find_one({"chat_id": old_chat_id})
        if old_chat:
            return await self.collection.update(
                {"chat_id": old_chat_id},
                {"chat_id": new_chat_id},
            )
        return
