from contextlib import suppress
from hashlib import md5
from threading import RLock
from time import time

from alita import LOGGER
from alita.database import MongoDB
from alita.utils.msg_types import Types

INSERTION_LOCK = RLock()
NOTES_CACHE = {}


class Notes(MongoDB):
    db_name = "notes"

    def __init__(self) -> None:
        super().__init__(self.db_name)

    def save_note(
        self,
        chat_id: int,
        note_name: str,
        note_value: str,
        msgtype: int = Types.TEXT,
        fileid="",
    ):
        with INSERTION_LOCK:

            # local dict update
            try:
                curr_notes = NOTES_CACHE[chat_id]
            except KeyError:
                curr_notes = []

            try:
                keywords = {i["note_name"] for i in curr_notes}
            except KeyError:
                keywords = set()
            hash_gen = md5(
                (note_name + note_value + str(chat_id) + str(int(time()))).encode(),
            ).hexdigest()

            if note_name not in keywords:
                curr_notes.append(
                    {
                        "chat_id": chat_id,
                        "note_name": note_name,
                        "note_value": note_value,
                        "hash": hash_gen,
                        "msgtype": msgtype,
                        "fileid": fileid,
                    },
                )
                NOTES_CACHE[chat_id] = curr_notes

            if note_name in keywords:
                with suppress(KeyError, StopIteration):
                    curr_notes.remove(
                        next(i for i in curr_notes if note_name == i["note_name"]),
                    )
                curr_notes.append(
                    {
                        "chat_id": chat_id,
                        "note_name": note_name,
                        "note_value": note_value,
                        "hash": hash_gen,
                        "msgtype": msgtype,
                        "fileid": fileid,
                    },
                )
                NOTES_CACHE[chat_id] = curr_notes

            if self.find_one(
                {"chat_id": chat_id, "note_name": note_name},
            ):
                self.update(
                    {"chat_id": chat_id},
                    {
                        "note_value": note_value,
                        "hash": hash_gen,
                        "msgtype": msgtype,
                        "fileid": fileid,
                    },
                )
                return True
            return self.insert_one(
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
            with suppress(KeyError, StopIteration):
                if curr := next(
                    i for i in NOTES_CACHE[chat_id] if note_name == i["note_name"]
                ):
                    return curr
            if curr := self.find_one(
                {"chat_id": chat_id, "note_name": note_name},
            ):
                return curr
            return "Note does not exist!"

    def get_note_by_hash(self, note_hash: str):
        return self.find_one({"hash": note_hash})

    def get_all_notes(self, chat_id: int):
        with INSERTION_LOCK:
            with suppress(KeyError):
                return [(i["note_name"], i["hash"]) for i in NOTES_CACHE[chat_id]]
            curr = self.find_all({"chat_id": chat_id})
            note_list = [(note["note_name"], note["hash"]) for note in curr]
            note_list.sort()
            return note_list

    def rm_note(self, chat_id: int, note_name: str):
        with INSERTION_LOCK:
            with suppress(KeyError, StopIteration):
                NOTES_CACHE[chat_id].remove(
                    next(
                        i for i in NOTES_CACHE[chat_id] if note_name == i["note_name"]
                    ),
                )
            if curr := self.find_one(
                {"chat_id": chat_id, "note_name": note_name},
            ):
                self.delete_one(curr)
                return True
            return False

    def rm_all_notes(self, chat_id: int):
        with INSERTION_LOCK:
            with suppress(KeyError):
                del NOTES_CACHE[chat_id]
            return self.delete_one({"chat_id": chat_id})

    def count_notes(self, chat_id: int):
        with INSERTION_LOCK:
            return len(curr) if (curr := self.find_all({"chat_id": chat_id})) else 0

    def count_notes_chats(self):
        with INSERTION_LOCK:
            notes = self.find_all()
            chats_ids = [chat["chat_id"] for chat in notes]
            return len(set(chats_ids))

    def count_all_notes(self):
        with INSERTION_LOCK:
            return self.count()

    def count_notes_type(self, ntype):
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
                NOTES_CACHE[new_chat_id] = NOTES_CACHE.pop(old_chat_id)

            if old_chat_db := self.find_one({"_id": old_chat_id}):
                new_data = old_chat_db.update({"_id": new_chat_id})
                self.delete_one({"_id": old_chat_id})
                self.insert_one(new_data)


def __pre_req_notes():
    global NOTES_CACHE
    start = time()
    db = Notes()
    all_notes = db.load_from_db()

    chat_ids = {i["chat_id"] for i in all_notes}

    for i in all_notes:
        del i["_id"]

    NOTES_CACHE = {
        chat: [filt for filt in all_notes if filt["chat_id"] == chat]
        for chat in chat_ids
    }
    LOGGER.info(f"Loaded Notes Cache - {round((time() - start), 3)}s")


class NotesSettings(MongoDB):
    db_name = "notes_settings"

    def __init__(self) -> None:
        super().__init__(self.db_name)

    def set_privatenotes(self, chat_id: int, status: bool = False):
        if self.find_one({"_id": chat_id}):
            return self.update({"_id": chat_id}, {"privatenotes": status})
        return self.insert_one({"_id": chat_id, "privatenotes": status})

    def get_privatenotes(self, chat_id: int):
        if curr := self.find_one({"_id": chat_id}):
            return curr["privatenotes"]
        self.update({"_id": chat_id}, {"privatenotes": False})
        return False

    def list_chats(self):
        return self.find_all({"privatenotes": True})

    def count_chats(self):
        return len(self.find_all({"privatenotes": True}))

    # Migrate if chat id changes!
    def migrate_chat_pnotes(self, old_chat_id: int, new_chat_id: int):
        with INSERTION_LOCK:
            if old_chat_db := self.find_one({"_id": old_chat_id}):
                new_data = old_chat_db.update({"_id": new_chat_id})
                self.delete_one({"_id": old_chat_id})
                self.insert_one(new_data)
