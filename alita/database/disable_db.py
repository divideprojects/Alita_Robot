from threading import RLock
from time import time

from alita import LOGGER
from alita.database import MongoDB

INSERTION_LOCK = RLock()
DISABLED_CMDS = {}


class Disabling(MongoDB):
    """Class to manage database for Disabling for chats."""

    # Database name to connect to perform operations
    db_name = "disabled"

    def __init__(self, chat_id: int) -> None:
        super().__init__(self.db_name)
        self.chat_id = chat_id
        self.chat_info = self.__ensure_in_db()

    def check_cmd_status(self, cmd: str):
        with INSERTION_LOCK:
            try:
                cmds = DISABLED_CMDS[self.chat_id]["commands"]
            except KeyError:
                cmds = self.chat_info["commands"]
                act = self.chat_info["action"]
                DISABLED_CMDS[self.chat_id] = {
                    "command": cmds if cmds else [],
                    "action": act if act else "none",
                }
            # return bool(cmd in cmds)
            return bool(cmd in cmds if cmds else [])

    def add_disable(self, cmd: str):
        with INSERTION_LOCK:
            if not self.check_cmd_status(cmd):
                return self.update(
                    {"_id": self.chat_id},
                    {
                        "_id": self.chat_id,
                        "commands": self.chat_info["commands"].append(cmd),
                    },
                )

    def remove_disabled(self, comm: str):
        with INSERTION_LOCK:
            if self.check_cmd_status(comm):
                self.chat_info["commands"].remove(comm)
                return self.update(
                    {"_id": self.chat_id},
                    {
                        "_id": self.chat_id,
                        "commands": self.chat_info["commands"],
                    },
                )

    def get_disabled(self):
        with INSERTION_LOCK:
            try:
                cmds = DISABLED_CMDS[self.chat_id]["commands"]
            except KeyError:
                cmds = self.chat_info["commands"]
                DISABLED_CMDS[self.chat_id] = {
                    "commands": cmds if cmds else [],
                    "action": self.chat_info["action"],
                }
            return cmds if cmds else []

    @staticmethod
    def count_disabled_all():
        with INSERTION_LOCK:
            collection = MongoDB(Disabling.db_name)
            curr = collection.find_all()
            return sum(
                len(chat["commands"] if chat["commands"] else [])
                for chat in curr)

    @staticmethod
    def count_disabling_chats():
        with INSERTION_LOCK:
            collection = MongoDB(Disabling.db_name)
            curr = collection.find_all()
            return sum(1 for chat in curr if chat["commands"])

    def set_action(self, action: str):
        with INSERTION_LOCK:
            try:
                DISABLED_CMDS[self.chat_id]["action"] = action
            except KeyError:
                cmds = self.chat_info["commands"]
                DISABLED_CMDS[self.chat_id] = {
                    "commands": cmds if cmds else [],
                    "action": action,
                }
            return self.update(
                {"_id": self.chat_id},
                {
                    "_id": self.chat_id,
                    "action": action
                },
            )

    def get_action(self):
        with INSERTION_LOCK:
            try:
                val = DISABLED_CMDS[self.chat_id]["action"]
            except KeyError:
                cmds = self.chat_info["commands"]
                val = self.chat_info["action"]
                DISABLED_CMDS[self.chat_id] = {
                    "commands": cmds if cmds else [],
                    "action": val,
                }
            return val if val else "none"

    @staticmethod
    def count_action_dis_all(action: str):
        with INSERTION_LOCK:
            collection = MongoDB(Disabling.db_name)
            all_data = collection.find_all({"action": action})
            return sum(
                len(i["commands"] if i["commands"] else []) >= 1
                for i in all_data)

    def rm_all_disabled(self):
        with INSERTION_LOCK:
            try:
                DISABLED_CMDS[self.chat_id]["commands"] = []
            except KeyError:
                DISABLED_CMDS[self.chat_id] = {
                    "commands": [],
                    "action": self.chat_info["action"],
                }
            return self.update(
                {"_id": self.chat_id},
                {"commands": []},
            )

    def __ensure_in_db(self):
        try:
            chat_data = DISABLED_CMDS[self.chat_id]
        except KeyError:
            chat_data = self.find_one({"_id": self.chat_id})
            if not chat_data:
                new_data = {
                    "_id": self.chat_id,
                    "commands": [],
                    "action": "none",
                }
                DISABLED_CMDS[self.chat_id] = {
                    "commands": [],
                    "action": "none"
                }
                self.insert_one(new_data)
                LOGGER.info(
                    f"Initialized Disabling Document for chat {self.chat_id}")
                return new_data
            DISABLED_CMDS[self.chat_id] = chat_data
        return chat_data

    # Migrate if chat id changes!
    def migrate_chat(self, new_chat_id: int):
        old_chat_db = self.find_one({"_id": self.chat_id})
        new_data = old_chat_db.update({"_id": new_chat_id})
        DISABLED_CMDS[new_chat_id] = DISABLED_CMDS.pop(self.chat_id)
        self.insert_one(new_data)
        self.delete_one({"_id": self.chat_id})

    @staticmethod
    def repair_db(collection):
        global DISABLED_CMDS
        all_data = collection.find_all()
        DISABLED_CMDS = {
            i["_id"]: {
                "action": i["action"] if i["action"] else "none",
                "commands": i["commands"] if i["commands"] else [],
            }
            for i in all_data
        }
        keys = {
            "commands": [],
            "action": "none",
        }
        for data in all_data:
            for key, val in keys.items():
                try:
                    _ = data[key]
                except KeyError:
                    LOGGER.warning(
                        f"Repairing Disabling Database - setting '{key}:{val}' for {data['_id']}",
                    )
                    collection.update({"_id": data["_id"]}, {key: val})


def __pre_req_disabling():
    start = time()
    LOGGER.info("Starting disabling Database Repair ...")
    collection = MongoDB(Disabling.db_name)
    Disabling.repair_db(collection)
    LOGGER.info(f"Done in {round((time() - start), 3)}s!")


__pre_req_disabling()
