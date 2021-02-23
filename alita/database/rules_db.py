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


class Rules:
    """Class for rules for chats in bot."""

    def __init__(self) -> None:
        self.collection = MongoDB("rules")

    async def get_rules(self, chat_id: int):
        rules = (await self.collection.find_one({"chat_id": chat_id}))["rules"]
        if rules:
            return rules
        return None

    async def set_rules(self, chat_id: int, rules: str):
        curr_rules = (await self.collection.find_one({"chat_id": chat_id}))["rules"]
        if curr_rules:
            return await self.collection.replace(
                {"chat_id": chat_id},
                {"chat_id": chat_id, "rules": rules},
            )
        return await self.collection.insert_one({"chat_id": chat_id, "rules": rules})

    async def clear_rules(self, chat_id: int):
        curr_rules = (await self.collection.find_one({"chat_id": chat_id}))["rules"]
        if curr_rules:
            return await self.collection.delete_one({"chat_id": chat_id})
        return
