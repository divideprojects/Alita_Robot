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

from alita import LOGGER
from alita.database import MongoDB

INSERTION_LOCK = RLock()
DISABLED_CMDS = {}


class Disabling(MongoDB):
    """Class to manage database for Disabling for chats."""

    # Database name to connect to to preform operations
    db_name = "disabled"

    def __init__(self, chat_id: int) -> None:
        super().__init__(self.db_name)
        self.chat_id = chat_id
        self.chat_info = self.__ensure_in_db()

    def check_cmd_status(self, cmd: str):
        with INSERTION_LOCK:
            # cmds = self.chat_info["commands"]
            cmds = DISABLED_CMDS[self.chat_id]["commands"]
            # return bool(cmd in cmds)
            return bool(cmd in cmds)

    def add_disable(self, cmd: str):
        with INSERTION_LOCK:
            if not self.check_cmd_status(cmd):
                DISABLED_CMDS[self.chat_id]["commands"].append(cmd)
                return self.update(
                    {"_id": self.chat_id},
                    {
                        "_id": self.chat_id,
                        "commands": self.chat_info["commands"] + [cmd],
                    },
                )

    def remove_disabled(self, comm: str):
        with INSERTION_LOCK:
            if self.check_cmd_status(comm):
                self.chat_info["commands"].remove(comm)
                DISABLED_CMDS[self.chat_id]["commands"].remove(comm)
                return self.update(
                    {"_id": self.chat_id},
                    {
                        "_id": self.chat_id,
                        "commands": self.chat_info["commands"],
                    },
                )

    def get_disabled(self):
        with INSERTION_LOCK:
            global DISABLED_CMDS
            try:
                cmds = DISABLED_CMDS[self.chat_id]["commands"]
            except KeyError:
                cmds = self.chat_info["commands"]
                DISABLED_CMDS[self.chat_id]["commands"] = cmds
            return cmds

    @staticmethod
    def count_disabled_all():
        with INSERTION_LOCK:
            collection = MongoDB(Disabling.db_name)
            curr = collection.find_all()
            return sum(len(chat["commands"]) for chat in curr)

    @staticmethod
    def count_disabling_chats():
        with INSERTION_LOCK:
            collection = MongoDB(Disabling.db_name)
            curr = collection.find_all()
            return sum(1 for chat in curr if chat["commands"])

    def set_action(self, action: str):
        with INSERTION_LOCK:
            global DISABLED_CMDS
            DISABLED_CMDS[self.chat_id]["action"] = action
            return self.update(
                {"_id": self.chat_id},
                {"_id": self.chat_id, "action": action},
            )

    def get_action(self):
        with INSERTION_LOCK:
            global DISABLED_CMDS
            try:
                action = DISABLED_CMDS[self.chat_id]["action"]
            except KeyError:
                action = self.chat_info["action"]
                DISABLED_CMDS[self.chat_id]["action"] = action
            return action

    @staticmethod
    def count_action_dis_all(action: str):
        with INSERTION_LOCK:
            collection = MongoDB(Disabling.db_name)
            all_data = collection.find_all({"action": action})
            return sum(len(i["commands"]) >= 1 for i in all_data)

    def rm_all_disabled(self):
        with INSERTION_LOCK:
            DISABLED_CMDS[self.chat_id]["commands"] = []
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
                new_data = new_data = {
                    "_id": self.chat_id,
                    "commands": [],
                    "action": "none",
                }
                self.insert_one(new_data)
                LOGGER.info(f"Initialized Disabling Document for chat {self.chat_id}")
                return new_data
        return chat_data

    # Migrate if chat id changes!
    def migrate_chat(self, new_chat_id: int):
        global DISABLED_CMDS  # global only when we are modifying the value
        old_chat_db = self.find_one({"_id": self.chat_id})
        new_data = old_chat_db.update({"_id": new_chat_id})
        DISABLED_CMDS[new_chat_id] = DISABLED_CMDS[self.chat_id]
        del DISABLED_CMDS[self.chat_id]
        self.insert_one(new_data)
        self.delete_one({"_id": self.chat_id})


def __load_disable_cache():
    global DISABLED_CMDS
    collection = MongoDB(Disabling.db_name)
    all_data = collection.find_all()
    DISABLED_CMDS = {
        i["_id"]: {"action": i["action"], "commands": i["commands"]} for i in all_data
    }


__load_disable_cache()
