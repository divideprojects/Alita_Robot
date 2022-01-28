from threading import RLock

from alita.database import MongoDB
from alita.database.chats_db import Chats

INSERTION_LOCK = RLock()
BLACKLIST_CHATS = []


class GroupBlacklist(MongoDB):
    """Class to blacklist chats where bot will exit."""

    db_name = "group_blacklists"

    def __init__(self) -> None:
        super().__init__(self.db_name)

    def add_chat(self, chat_id: int):
        with INSERTION_LOCK:
            global BLACKLIST_CHATS
            try:
                Chats.remove_chat(chat_id)  # Delete chat from database
            except KeyError:
                pass
            BLACKLIST_CHATS.append(chat_id)
            BLACKLIST_CHATS.sort()
            return self.insert_one({"_id": chat_id, "blacklist": True})

    def remove_chat(self, chat_id: int):
        with INSERTION_LOCK:
            global BLACKLIST_CHATS
            BLACKLIST_CHATS.remove(chat_id)
            BLACKLIST_CHATS.sort()
            return self.delete_one({"_id": chat_id})

    def list_all_chats(self):
        with INSERTION_LOCK:
            try:
                BLACKLIST_CHATS.sort()
                return BLACKLIST_CHATS
            except Exception:
                all_chats = self.find_all()
                return [chat["_id"] for chat in all_chats]

    def get_from_db(self):
        return self.find_all()
