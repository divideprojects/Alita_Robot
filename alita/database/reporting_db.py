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


class Reporting:
    """Class for managing report settings of users and groups."""

    def __init__(self) -> None:
        self.collection = MongoDB("reporting_settings")

    async def get_chat_type(chat_id: int):
        if str(chat_id).startswith("-100"):
            chat_type = "supergroup"
        else:
            chat_type = "user"
        return chat_type

    async def set_settings(self, chat_id: int, status: bool = True):
        chat_type = await self.get_chat_type(chat_id)
        curr_settings = (await self.collection.find_one({"chat_id": chat_id}))["status"]
        if curr_settings:
            return await self.collection.update(
                {"chat_id": chat_id},
                {"status": status},
            )
        return await self.collection.insert_one(
            {"chat_id": chat_id, "chat_type": chat_type, "status": status},
        )

    async def get_settings(self, chat_id: int):
        chat_type = await self.get_chat_type(chat_id)
        curr_settings = (await self.collection.find_one({"chat_id": chat_id}))["status"]
        if curr_settings:
            return curr_settings
        await self.collection.insert_one(
            {"chat_id": chat_id, "chat_type": chat_type, "status": True},
        )
        return True
