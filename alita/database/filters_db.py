from contextlib import suppress
from threading import RLock
from time import time
from traceback import format_exc

from alita import LOGGER
from alita.database import MongoDB
from alita.utils.msg_types import Types

INSERTION_LOCK = RLock()
FILTER_CACHE = {}


class Filters(MongoDB):
    db_name = "chat_filters"

    def __init__(self) -> None:
        super().__init__(self.db_name)

    def save_filter(
        self,
        chat_id: int,
        keyword: str,
        filter_reply: str,
        msgtype: int = Types.TEXT,
        fileid="",
    ):
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

            if keyword in keywords:
                with suppress(KeyError, StopIteration):
                    curr_filters.remove(
                        next(
                            i
                            for i in curr_filters
                            if keyword in i["keyword"].split("|")
                        ),
                    )
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

            if self.find_one(
                {"chat_id": chat_id, "keyword": keyword},
            ):
                self.update(
                    {"chat_id": chat_id},
                    {
                        "filter_reply": filter_reply,
                        "msgtype": msgtype,
                        "fileid": fileid,
                    },
                )
                return True
            return self.insert_one(
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
            with suppress(KeyError, StopIteration):
                if curr := next(
                    i
                    for i in FILTER_CACHE[chat_id]
                    if keyword in i["keyword"].split("|")
                ):
                    return curr

            if curr := self.find_one(
                {"chat_id": chat_id, "keyword": {"$regex": rf"\|?{keyword}\|?"}},
            ):
                return curr

            return "Filter does not exist!"

    def get_all_filters(self, chat_id: int):
        with INSERTION_LOCK:
            with suppress(KeyError):
                return [i["keyword"] for i in FILTER_CACHE[chat_id]]

            if curr := self.find_all({"chat_id": chat_id}):
                filter_list = {i["keyword"] for i in curr}
                return list(filter_list)
            return []

    def rm_filter(self, chat_id: int, keyword: str):
        with INSERTION_LOCK:
            with suppress(KeyError, StopIteration):
                FILTER_CACHE[chat_id].remove(
                    next(
                        i
                        for i in FILTER_CACHE[chat_id]
                        if keyword in i["keyword"].split("|")
                    ),
                )

            if curr := self.find_one(
                {"chat_id": chat_id, "keyword": {"$regex": rf"\|?{keyword}\|?"}},
            ):
                self.delete_one(curr)
                return 1

            return 0

    def rm_all_filters(self, chat_id: int):
        with INSERTION_LOCK:
            with suppress(KeyError):
                del FILTER_CACHE[chat_id]

            return self.delete_one({"chat_id": chat_id})

    def count_filters_all(self):
        with INSERTION_LOCK:
            with suppress(KeyError):
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

            # curr = self.find_all()
            # if curr:
            #    return len(curr)
            return self.count()

    def count_filter_aliases(self):
        with INSERTION_LOCK:
            with suppress(KeyError):
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

            if curr := self.find_all():
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
            filters = self.find_all()
            chats_ids = {i["chat_id"] for i in filters}
            return len(chats_ids)

    def count_all_filters(self):
        with INSERTION_LOCK:
            with suppress(KeyError):
                return len(
                    [
                        (i["keyword"] for i in FILTER_CACHE[chat_id])
                        for chat_id in set(FILTER_CACHE.keys())
                    ],
                )
            return self.count()

    def count_filter_type(self, ntype):
        with INSERTION_LOCK:
            return self.count({"msgtype": ntype})

    def load_from_db(self):
        with INSERTION_LOCK:
            return self.find_all()

    # Migrate if chat id changes!
    def migrate_chat(self, old_chat_id: int, new_chat_id: int):
        with INSERTION_LOCK:
            # Update locally
            with suppress(KeyError):
                FILTER_CACHE[new_chat_id] = FILTER_CACHE.pop(old_chat_id)

            if old_chat_db := self.find_one({"_id": old_chat_id}):
                new_data = old_chat_db.update({"_id": new_chat_id})
                self.delete_one({"_id": old_chat_id})
                self.insert_one(new_data)


def __pre_req_filters():
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
    LOGGER.info(f"Loaded Filters Cache - {round((time() - start), 3)}s")
