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


from hashlib import md5
from threading import RLock
from time import time

from alita.database import MongoDB
from alita.utils.msg_types import Types

INSERTION_LOCK = RLock()


class Notes:
    def __init__(self) -> None:
        self.collection = MongoDB("notes")

    def save_note(
        self,
        chat_id: int,
        note_name: str,
        note_value: str,
        msgtype: int = Types.TEXT,
        fileid="",
    ):
        with INSERTION_LOCK:
            curr = self.collection.find_one(
                {"chat_id": chat_id, "note_name": note_name},
            )
            if curr:
                return False
            hash_gen = md5(
                (note_name + note_value + str(chat_id) + str(int(time()))).encode(),
            ).hexdigest()
            return self.collection.insert_one(
                {
                    "chat_id": chat_id,
                    "note_name": note_name,
                    "note_value": note_value,
                    "hash": hash_gen,
                    "msgtype": msgtype,
                    "fileid": fileid,
                },
            )

    def get_note(self, chat_id: int, note_name: str):
        with INSERTION_LOCK:
            curr = self.collection.find_one(
                {"chat_id": chat_id, "note_name": note_name},
            )
            if curr:
                return curr
            return "Note does not exist!"

    def get_note_by_hash(self, note_hash: str):
        return self.collection.find_one({"hash": note_hash})

    def get_all_notes(self, chat_id: int):
        with INSERTION_LOCK:
            curr = self.collection.find_all({"chat_id": chat_id})
            note_list = []
            for note in curr:
                note_list.append((note["note_name"], note["hash"]))
            note_list.sort()
            return note_list

    def rm_note(self, chat_id: int, note_name: str):
        with INSERTION_LOCK:
            curr = self.collection.find_one(
                {"chat_id": chat_id, "note_name": note_name},
            )
            if curr:
                self.collection.delete_one(curr)
                return True
            return False

    def rm_all_notes(self, chat_id: int):
        with INSERTION_LOCK:
            return self.collection.delete_one({"chat_id": chat_id})

    def count_notes(self, chat_id: int):
        with INSERTION_LOCK:
            curr = self.collection.find_all({"chat_id": chat_id})
            if curr:
                return len(curr)
            return 0

    def count_notes_chats(self):
        with INSERTION_LOCK:
            notes = self.collection.find_all()
            chats_ids = []
            for chat in notes:
                chats_ids.append(chat["chat_id"])
            return len(set(chats_ids))

    def count_all_notes(self):
        with INSERTION_LOCK:
            return self.collection.count()

    def count_notes_type(self, ntype):
        with INSERTION_LOCK:
            return self.collection.count({"msgtype": ntype})

    # Migrate if chat id changes!
    def migrate_chat(self, old_chat_id: int, new_chat_id: int):
        with INSERTION_LOCK:

            old_chat_db = self.collection.find_one({"_id": old_chat_id})
            if old_chat_db:
                new_data = old_chat_db.update({"_id": new_chat_id})
                self.collection.delete_one({"_id": old_chat_id})
                self.collection.insert_one(new_data)


class NotesSettings:
    def __init__(self) -> None:
        self.collection = MongoDB("notes_settings")

    def set_privatenotes(self, chat_id: int, status: bool = False):
        curr = self.collection.find_one({"_id": chat_id})
        if curr:
            return self.collection.update({"_id": chat_id}, {"privatenotes": status})
        return self.collection.insert_one({"_id": chat_id, "privatenotes": status})

    def get_privatenotes(self, chat_id: int):
        curr = self.collection.find_one({"_id": chat_id})
        if curr:
            return curr["privatenotes"]
        self.collection.update({"_id": chat_id}, {"privatenotes": False})
        return False

    def list_chats(self):
        return self.collection.find_all({"privatenotes": True})

    def count_chats(self):
        return len(self.collection.find_all({"privatenotes": True}))

    # Migrate if chat id changes!
    def migrate_chat(self, old_chat_id: int, new_chat_id: int):
        with INSERTION_LOCK:

            old_chat_db = self.collection.find_one({"_id": old_chat_id})
            if old_chat_db:
                new_data = old_chat_db.update({"_id": new_chat_id})
                self.collection.delete_one({"_id": old_chat_id})
                self.collection.insert_one(new_data)
