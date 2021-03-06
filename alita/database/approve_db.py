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
            {"chat_id": chat_id},
        )
        if curr_approve:
            st = True if user_id in curr_approve["users"] else False
            return st
        return False

    async def add_approve(self, chat_id: int, user_id: int):
        curr = await self.collection.find_one({"chat_id": chat_id})
        if curr:
            users_old = curr["users"]
            users_old.append(user_id)
            users = list(dict.fromkeys(users_old))
            return await self.collection.update(
                {"chat_id": chat_id},
                {
                    "chat_id": chat_id,
                    "users": users,
                },
            )
        return await self.collection.insert_one(
            {
                "chat_id": chat_id,
                "users": [user_id],
            },
        )

    async def remove_approve(self, chat_id: int, user_id: int):
        curr = await self.collection.find_one({"chat_id": chat_id})
        if curr:
            users = curr["users"]
            users.remove(user_id)
            return await self.collection.update(
                {"chat_id": chat_id},
                {
                    "chat_id": chat_id,
                    "users": users,
                },
            )
        return "Not approved"

    async def unapprove_all(self, chat_id: int):
        return await self.collection.delete_one(
            {"chat_id": chat_id},
        )

    async def list_approved(self, chat_id: int):
        return ((await self.collection.find_all({"chat_id": chat_id}))["users"]) or []

    async def count_all_approved(self):
        num = 0
        curr = await self.collection.find_all()
        if curr:
            for chat in curr:
                users = chat["users"]
                num += len(users)

        return num

    async def count_approved_chats(self):
        return (await self.collection.count()) or 0

    async def count_approved(self, chat_id: int):
        all_app = await self.collection.find_one({"chat_id": chat_id})
        return len(all_app["users"]) or 0

    # Migrate if chat id changes!
    async def migrate_chat(self, old_chat_id: int, new_chat_id: int):
        old_chat = await self.collection.find_one({"chat_id": old_chat_id})
        if old_chat:
            return await self.collection.update(
                {"chat_id": old_chat_id},
                {"chat_id": new_chat_id},
            )
        return
