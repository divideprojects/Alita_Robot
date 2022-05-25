from threading import RLock
from time import time

from alita import LOGGER
from alita.database import MongoDB

INSERTION_LOCK = RLock()


class Chats(MongoDB):
    """Class to manage users for bot."""

    # Database name to connect to to preform operations
    db_name = "chats"

    def __init__(self, chat_id: int) -> None:
        super().__init__(self.db_name)
        self.chat_id = chat_id
        self.chat_info = self.__ensure_in_db()

    def user_is_in_chat(self, user_id: int):
        return user_id in set(self.chat_info["users"])

    def update_chat(self, chat_name: str, user_id: int):
        with INSERTION_LOCK:

            if chat_name == self.chat_info["chat_name"] and self.user_is_in_chat(
                user_id,
            ):
                return True

            if chat_name != self.chat_info["chat_name"] and self.user_is_in_chat(
                user_id,
            ):
                return self.update(
                    {"_id": self.chat_id},
                    {"chat_name": chat_name},
                )

            if chat_name == self.chat_info["chat_name"] and not self.user_is_in_chat(
                user_id,
            ):
                self.chat_info["users"].append(user_id)
                return self.update(
                    {"_id": self.chat_id},
                    {"users": self.chat_info["users"]},
                )

            users_old = self.chat_info["users"]
            users_old.append(user_id)
            users = list(set(users_old))
            return self.update(
                {"_id": self.chat_id},
                {
                    "_id": self.chat_id,
                    "chat_name": chat_name,
                    "users": users,
                },
            )

    def count_chat_users(self):
        with INSERTION_LOCK:
            return len(self.chat_info["users"]) or 0

    def chat_members(self):
        with INSERTION_LOCK:
            return self.chat_info["users"]

    @staticmethod
    def remove_chat(chat_id: int):
        with INSERTION_LOCK:
            collection = MongoDB(Chats.db_name)
            collection.delete_one({"_id": chat_id})

    @staticmethod
    def count_chats():
        with INSERTION_LOCK:
            collection = MongoDB(Chats.db_name)
            return collection.count() or 0

    @staticmethod
    def list_chats_by_id():
        with INSERTION_LOCK:
            collection = MongoDB(Chats.db_name)
            chats = collection.find_all()
            chat_list = {i["_id"] for i in chats}
            return list(chat_list)

    @staticmethod
    def list_chats_full():
        with INSERTION_LOCK:
            collection = MongoDB(Chats.db_name)
            return collection.find_all()

    @staticmethod
    def get_chat_info(chat_id: int):
        with INSERTION_LOCK:
            collection = MongoDB(Chats.db_name)
            return collection.find_one({"_id": chat_id})

    def load_from_db(self):
        with INSERTION_LOCK:
            return self.find_all()

    def __ensure_in_db(self):
        chat_data = self.find_one({"_id": self.chat_id})
        if not chat_data:
            new_data = {"_id": self.chat_id, "chat_name": "", "users": []}
            self.insert_one(new_data)
            LOGGER.info(f"Initialized Chats Document for chat {self.chat_id}")
            return new_data
        return chat_data

    # Migrate if chat id changes!
    def migrate_chat(self, new_chat_id: int):
        old_chat_db = self.find_one({"_id": self.chat_id})
        new_data = old_chat_db.update({"_id": new_chat_id})
        self.insert_one(new_data)
        self.delete_one({"_id": self.chat_id})

    @staticmethod
    def repair_db(collection):
        all_data = collection.find_all()
        keys = {"chat_name": "", "users": []}
        for data in all_data:
            for key, val in keys.items():
                try:
                    _ = data[key]
                except KeyError:
                    LOGGER.warning(
                        f"Repairing Chats Database - setting '{key}:{val}' for {data['_id']}",
                    )
                    collection.update({"_id": data["_id"]}, {key: val})


def __pre_req_chats():
    start = time()
    LOGGER.info("Starting Chats Database Repair...")
    collection = MongoDB(Chats.db_name)
    Chats.repair_db(collection)
    LOGGER.info(f"Done in {round((time() - start), 3)}s!")
