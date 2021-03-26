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
from time import time
from traceback import format_exc

from alita import LOGGER
from alita.database import MongoDB
from alita.utils.msg_types import Types

INSERTION_LOCK = RLock()

FILTER_CACHE = {}


class Filters:
    def __init__(self) -> None:
        self.collection = MongoDB("chat_filters")

    def save_filter(
        self,
        chat_id: int,
        keyword: str,
        filter_reply: str,
        msgtype: int = Types.TEXT,
        fileid="",
    ):
        global FILTER_CACHE
        with INSERTION_LOCK:

            # local dict update
            try:
                curr_filters = FILTER_CACHE[chat_id]
            except KeyError:
                curr_filters = []

            try:
                keywords = {i["keyword"] for i in curr_filters}
            except KeyError:
                keywords = set()

            if keyword not in keywords:
                curr_filters.append(
                    {
                        "chat_id": chat_id,
                        "keyword": keyword,
                        "filter_reply": filter_reply,
                        "msgtype": msgtype,
                        "fileid": fileid,
                    },
                )
                FILTER_CACHE[chat_id] = curr_filters

            # Database update
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
            try:
                curr = next(
                    i
                    for i in FILTER_CACHE[chat_id]
                    if keyword in i["keyword"].split("|")
                )
                if curr:
                    return curr
            except (KeyError, StopIteration):
                pass
            except Exception as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())

            curr = self.collection.find_one(
                {"chat_id": chat_id, "keyword": {"$regex": fr"\|?{keyword}\|?"}},
            )
            if curr:
                return curr

            return "Filter does not exist!"

    def get_all_filters(self, chat_id: int):
        with INSERTION_LOCK:
            try:
                return [i["keyword"] for i in FILTER_CACHE[chat_id]]
            except KeyError:
                pass
            except Exception as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())

            curr = self.collection.find_all({"chat_id": chat_id})
            if curr:
                filter_list = {i["keyword"] for i in curr}
                return list(filter_list)
            return []

    def rm_filter(self, chat_id: int, keyword: str):
        global FILTER_CACHE
        with INSERTION_LOCK:
            try:
                FILTER_CACHE[chat_id].remove(
                    next(
                        i
                        for i in FILTER_CACHE[chat_id]
                        if keyword in i["keyword"].split("|")
                    ),
                )
            except KeyError:
                pass
            except Exception as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())

            curr = self.collection.find_one(
                {"chat_id": chat_id, "keyword": {"$regex": fr"\|?{keyword}\|?"}},
            )
            if curr:
                self.collection.delete_one(curr)
                return True

            return False

    def rm_all_filters(self, chat_id: int):
        global FILTER_CACHE
        with INSERTION_LOCK:
            try:
                del FILTER_CACHE[chat_id]
            except KeyError:
                pass
            except Exception as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())

            return self.collection.delete_one({"chat_id": chat_id})

    def count_filters_all(self):
        with INSERTION_LOCK:
            try:
                return len(
                    [
                        j
                        for i in (
                            (i["keyword"] for i in FILTER_CACHE[chat_id])
                            for chat_id in set(FILTER_CACHE.keys())
                        )
                        for j in i
                    ],
                )
            except KeyError:
                pass
            except Exception as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())

            curr = self.collection.find_all()
            if curr:
                return len(curr)
            return 0

    def count_filter_aliases(self):
        with INSERTION_LOCK:
            try:
                return len(
                    [
                        i
                        for j in [
                            j.split("|")
                            for i in (
                                (i["keyword"] for i in FILTER_CACHE[chat_id])
                                for chat_id in set(FILTER_CACHE.keys())
                            )
                            for j in i
                            if len(j.split("|")) >= 2
                        ]
                        for i in j
                    ],
                )
            except KeyError:
                pass
            except Exception as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())

            curr = self.collection.find_all()
            if curr:
                return len(
                    [z for z in (i["keyword"].split("|") for i in curr) if len(z) >= 2],
                )
            return 0

    def count_filters_chats(self):
        with INSERTION_LOCK:
            try:
                return len(set(FILTER_CACHE.keys()))
            except Exception as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())
            filters = self.collection.find_all()
            chats_ids = {i["chat_id"] for i in filters}
            return len(chats_ids)

    def count_all_filters(self):
        with INSERTION_LOCK:
            try:
                return len(
                    [
                        (i["keyword"] for i in FILTER_CACHE[chat_id])
                        for chat_id in set(FILTER_CACHE.keys())
                    ],
                )
            except KeyError:
                pass
            except Exception as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())
            return self.collection.count()

    def count_filter_type(self, ntype):
        with INSERTION_LOCK:
            return self.collection.count({"msgtype": ntype})

    def load_from_db(self):
        with INSERTION_LOCK:
            return self.collection.find_all()

    # Migrate if chat id changes!
    def migrate_chat(self, old_chat_id: int, new_chat_id: int):
        with INSERTION_LOCK:
            # Update locally
            try:
                old_db_local = FILTER_CACHE[old_chat_id]
                del FILTER_CACHE[old_chat_id]
                FILTER_CACHE[new_chat_id] = old_db_local
            except KeyError:
                pass

            # Update in db
            old_chat_db = self.collection.find_one({"_id": old_chat_id})
            if old_chat_db:
                new_data = old_chat_db.update({"_id": new_chat_id})
                self.collection.delete_one({"_id": old_chat_id})
                self.collection.insert_one(new_data)


def __load_filters_cache():
    global FILTER_CACHE
    start = time()
    db = Filters()
    all_filters = db.load_from_db()

    chat_ids = {i["chat_id"] for i in all_filters}

    for i in all_filters:
        del i["_id"]

    FILTER_CACHE = {
        chat: [filt for filt in all_filters if filt["chat_id"] == chat]
        for chat in chat_ids
    }
    LOGGER.info(f"Loaded Filters Cache - {round((time()-start),3)}s")
