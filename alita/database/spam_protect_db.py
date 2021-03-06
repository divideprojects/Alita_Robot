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


class SpamProtect:
    """Class for managing Spam Protection settings of chats!"""

    def __init__(self) -> None:
        self.collection = MongoDB("spam_protect")

    async def get_cas_status(self, chat_id: int):
        curr = await self.collection.find_one({"chat_id": chat_id})
        if curr:
            stat = curr["cas"]
            return stat
        await self.collection.insert_one(
            {"chat_id": chat_id, "cas": False, "underattack": False},
        )
        return False

    async def set_cas_status(self, chat_id: int, status: bool = False):
        curr = await self.collection.find_one({"chat_id": chat_id})
        if curr:
            return await self.collection.update(
                {"chat_id": chat_id},
                {"chat_id": chat_id, "cas": status},
            )
        await self.collection.insert_one(
            {"chat_id": chat_id, "cas": status, "underattack": False},
        )
        return status

    async def get_attack_status(self, chat_id: int):
        curr = await self.collection.find_one({"chat_id": chat_id})
        if curr:
            stat = curr["underattack"]
            return stat
        await self.collection.insert_one(
            {"chat_id": chat_id, "cas": False, "underattack": False},
        )
        return False

    async def set_attack_status(self, chat_id: int, status: bool = False):
        curr = await self.collection.find_one({"chat_id": chat_id})
        if curr:
            return await self.collection.update(
                {"chat_id": chat_id},
                {"chat_id": chat_id, "underattack": status},
            )
        await self.collection.insert_one(
            {"chat_id": chat_id, "cas": False, "underattack": status},
        )
        return status

    async def get_cas_enabled_chats_num(self):
        curr = await self.collection.find_all()
        num = 0
        if curr:
            for chat in curr:
                if chat["cas"]:
                    num += 1
        return num

    async def get_attack_enabled_chats_num(self):
        curr = await self.collection.find_all()
        num = 0
        if curr:
            for chat in curr:
                if chat["underattack"]:
                    num += 1
        return num

    async def get_cas_enabled_chats(self):
        curr = await self.collection.find_all()
        lst = []
        if curr:
            for chat in curr:
                if chat["cas"]:
                    lst.append(chat["chat_id"])
        return lst

    async def get_attack_enabled_chats(self):
        curr = await self.collection.find_all()
        lst = []
        if curr:
            for chat in curr:
                if chat["underattack"]:
                    lst.append(chat["chat_id"])
        return lst
