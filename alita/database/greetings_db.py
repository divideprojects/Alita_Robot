from threading import RLock

from alita import LOGGER
from alita.database import MongoDB

INSERTION_LOCK = RLock()


class Greetings(MongoDB):
    """Class for managing antichannelpins in chats."""

    # Database name to connect to to preform operations
    db_name = "welcome_chats"

    def __init__(self, chat_id: int) -> None:
        super().__init__(self.db_name)
        self.chat_id = chat_id
        self.chat_info = self.__ensure_in_db()

    # Get settings from database
    def get_welcome_status(self):
        with INSERTION_LOCK:
            return self.chat_info["welcome"]

    def get_goodbye_status(self):
        with INSERTION_LOCK:
            return self.chat_info["goodbye"]

    def get_current_cleanservice_settings(self):
        with INSERTION_LOCK:
            return self.chat_info["cleanservice"]

    def get_current_cleanwelcome_settings(self):
        with INSERTION_LOCK:
            return self.chat_info["cleanwelcome"]

    def get_current_cleangoodbye_settings(self):
        with INSERTION_LOCK:
            return self.chat_info["cleangoodbye"]

    def get_welcome_text(self):
        with INSERTION_LOCK:
            return self.chat_info["welcome_text"]

    def get_goodbye_text(self):
        with INSERTION_LOCK:
            return self.chat_info["goodbye_text"]

    def get_current_cleanwelcome_id(self):
        with INSERTION_LOCK:
            return self.chat_info["cleanwelcome_id"]

    def get_current_cleangoodbye_id(self):
        with INSERTION_LOCK:
            return self.chat_info["cleangoodbye_id"]

    # Set settings in database
    def set_current_welcome_settings(self, status: bool):
        with INSERTION_LOCK:
            return self.update({"_id": self.chat_id}, {"welcome": status})

    def set_current_goodbye_settings(self, status: bool):
        with INSERTION_LOCK:
            return self.update({"_id": self.chat_id}, {"goodbye": status})

    def set_welcome_text(self, welcome_text: str):
        with INSERTION_LOCK:
            return self.update(
                {"_id": self.chat_id},
                {"welcome_text": welcome_text},
            )

    def set_goodbye_text(self, goodbye_text: str):
        with INSERTION_LOCK:
            return self.update(
                {"_id": self.chat_id},
                {"goodbye_text": goodbye_text},
            )

    def set_current_cleanservice_settings(self, status: bool):
        with INSERTION_LOCK:
            return self.update(
                {"_id": self.chat_id},
                {"cleanservice": status},
            )

    def set_current_cleanwelcome_settings(self, status: bool):
        with INSERTION_LOCK:
            return self.update(
                {"_id": self.chat_id},
                {"cleanwelcome": status},
            )

    def set_current_cleangoodbye_settings(self, status: bool):
        with INSERTION_LOCK:
            return self.update(
                {"_id": self.chat_id},
                {"cleangoodbye": status},
            )

    def set_cleanwlcm_id(self, status: int):
        with INSERTION_LOCK:
            return self.update(
                {"_id": self.chat_id},
                {"cleanwelcome_id": status},
            )

    def set_cleangoodbye_id(self, status: int):
        with INSERTION_LOCK:
            return self.update(
                {"_id": self.chat_id},
                {"cleangoodbye_id": status},
            )

    def __ensure_in_db(self):
        chat_data = self.find_one({"_id": self.chat_id})
        if not chat_data:
            new_data = {
                "_id": self.chat_id,
                "cleanwelcome": False,
                "cleanwelcome_id": None,
                "cleangoodbye_id": None,
                "cleangoodbye": False,
                "cleanservice": False,
                "goodbye_text": "Sad to see you leaving {first}.\nTake Care!",
                "welcome_text": "Hey {first}, welcome to {chatname}!",
                "welcome": True,
                "goodbye": True,
            }
            self.insert_one(new_data)
            LOGGER.info(f"Initialized Greetings Document for chat {self.chat_id}")
            return new_data
        return chat_data

    # Migrate if chat id changes!
    def migrate_chat(self, new_chat_id: int):
        old_chat_db = self.find_one({"_id": self.chat_id})
        new_data = old_chat_db.update({"_id": new_chat_id})
        self.insert_one(new_data)
        self.delete_one({"_id": self.chat_id})

    @staticmethod
    def count_chats(query: str):
        with INSERTION_LOCK:
            collection = MongoDB(Greetings.db_name)
            return collection.count({query: True})
