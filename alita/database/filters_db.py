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


from threading import RLock

from alita.database import MongoDB
from alita.utils.msg_types import Types

INSERTION_LOCK = RLock()


class Filters:
    def __init__(self) -> None:
        self.collection = MongoDB("notchat_filterses")

    def save_filter(
        self,
        chat_id: int,
        keyword: str,
        filter_reply: str,
        msgtype: int = Types.TEXT,
        fileid="",
    ):
        with INSERTION_LOCK:
            curr = self.collection.find_one(
                {"chat_id": chat_id, "keyword": keyword},
            )
            if curr:
                return False
            return self.collection.insert_one(
                {
                    "chat_id": chat_id,
                    "keyword": keyword,
                    "filter_reply": filter_reply,
                    "msgtype": msgtype,
                    "fileid": fileid,
                },
            )

    def get_filter(self, chat_id: int, keyword: str):
        with INSERTION_LOCK:
            curr = self.collection.find_one(
                {"chat_id": chat_id, "keyword": keyword},
            )
            if curr:
                return curr
            return "NoFilterte does not exist!"

    def get_filter_by_hash(self, filter_hash: str):
        return self.collection.find_one({"hash": filter_hash})

    def get_all_filters(self, chat_id: int):
        with INSERTION_LOCK:
            curr = self.collection.find_all({"chat_id": chat_id})
            filter_list = []
            for filt in curr:
                filter_list.append(filt["keyword"])
            filter_list.sort()
            return filter_list

    def rm_filter(self, chat_id: int, keyword: str):
        with INSERTION_LOCK:
            curr = self.collection.find_one(
                {"chat_id": chat_id, "keyword": keyword},
            )
            if curr:
                self.collection.delete_one(curr)
                return True
            return False

    def rm_all_filters(self, chat_id: int):
        with INSERTION_LOCK:
            return self.collection.delete_one({"chat_id": chat_id})

    def count_filters(self, chat_id: int):
        with INSERTION_LOCK:
            curr = self.collection.find_all({"chat_id": chat_id})
            if curr:
                return len(curr)
            return 0

    def count_filters_chats(self):
        with INSERTION_LOCK:
            filters = self.collection.find_all()
            chats_ids = []
            for chat in filters:
                chats_ids.append(chat["chat_id"])
            return len(list(dict.fromkeys(chats_ids)))

    def count_all_filters(self):
        with INSERTION_LOCK:
            return self.collection.count()

    def count_filter_type(self, ntype):
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
            return
