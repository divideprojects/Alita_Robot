from datetime import datetime
from threading import RLock

from alita.database import MongoDB

INSERTION_LOCK = RLock()
ANTISPAM_BANNED = set()


class GBan(MongoDB):
    """Class for managing Gbans in bot."""

    db_name = "gbans"

    def __init__(self) -> None:
        super().__init__(self.db_name)

    def check_gban(self, user_id: int):
        with INSERTION_LOCK:
            return bool(self.find_one({"_id": user_id}))

    def add_gban(self, user_id: int, reason: str, by_user: int):
        global ANTISPAM_BANNED
        with INSERTION_LOCK:
            # Check if  user is already gbanned or not
            if self.find_one({"_id": user_id}):
                return self.update_gban_reason(user_id, reason)

            # If not already gbanned, then add to gban
            time_rn = datetime.now()
            return self.insert_one(
                {
                    "_id": user_id,
                    "reason": reason,
                    "by": by_user,
                    "time": time_rn,
                },
            )

    def remove_gban(self, user_id: int):
        global ANTISPAM_BANNED
        with INSERTION_LOCK:
            # Check if  user is already gbanned or not
            if self.find_one({"_id": user_id}):
                return self.delete_one({"_id": user_id})

            return "User not gbanned!"

    def get_gban(self, user_id: int):
        if self.check_gban(user_id):
            if curr := self.find_one({"_id": user_id}):
                return True, curr["reason"]
        return False, ""

    def update_gban_reason(self, user_id: int, reason: str):
        with INSERTION_LOCK:
            return self.update(
                {"_id": user_id},
                {"reason": reason},
            )

    def count_gbans(self):
        with INSERTION_LOCK:
            return self.count()

    def load_from_db(self):
        with INSERTION_LOCK:
            return self.find_all()

    def list_gbans(self):
        with INSERTION_LOCK:
            return self.find_all()
