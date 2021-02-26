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
from alita.utils.msg_types import Types


class Notes:
    def __init__(self) -> None:
        self.collection = MongoDB("notes")

    async def save_note(
        self,
        chat_id: int,
        note_name: str,
        note_value: str,
        msgtype: int = Types.TEXT,
        file=None,
    ):
        curr = await self.collection.find_one({"chat_id": chat_id, "note_name": note_name})
        if curr:
            return False
        return await self.collection.insert_one(
            {
                "chat_id": chat_id,
                "note_name": note_name,
                "note_value": note_value,
                "msgtype": msgtype,
                "file": file,
            },
        )

    async def get_note(self, chat_id: int, note_name: str):
        curr = (
            await self.collection.find_one({"chat_id": chat_id, "note_name": note_name})
        )
        if curr:
            return curr["note_value"]
        return "Note does not exist!"

    async def get_all_notes(self, chat_id: int):
        curr = await self.collection.find_all({"chat_id": chat_id})
        note_list = []
        for note in curr:
            note_list.append(note["note_name"])
        note_list.sort()
        return note_list

    async def rm_note(self, chat_id: int, note_name: str):
        curr = await self.collection.find_one(
            {"chat_id": chat_id, "note_name": note_name},
        )
        if curr:
            await self.collection.delete_one(curr)
            return True
        return False

    async def rm_all_notes(self, chat_id: int):
        note_list = await self.collection.get_all_notes({"chat_id": chat_id})
        if note_list:
            for note in note_list:
                await self.collection.rm_note(note)
            return True
        return False

    async def count_notes(self, chat_id: int):
        curr = await self.collection.find_all({"chat_id": chat_id})
        if curr:
            return len(curr)
        return 0

    async def count_notes_chats(self):
        notes = await self.collection.find_all()
        chats_ids = []
        for chat in notes:
            chats_ids.append(chat["chat_id"])
        return len(list(dict.fromkeys(chats_ids)))

    async def count_all_notes(self):
        return len(await self.collection.find_all())
