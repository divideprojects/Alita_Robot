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


import threading

from sqlalchemy import Column, String, UnicodeText, distinct, func

from alita.database import BASE, SESSION


class Rules(BASE):
    __tablename__ = "rules"
    chat_id = Column(String(14), primary_key=True)
    rules = Column(UnicodeText, default="")

    def __init__(self, chat_id):
        self.chat_id = chat_id

    def __repr__(self):
        return f"<Chat {self.chat_id} rules: {self.rules}>"


Rules.__table__.create(checkfirst=True)

INSERTION_LOCK = threading.RLock()


def set_rules(chat_id, rules_text):
    with INSERTION_LOCK:
        try:
            rules = SESSION.query(Rules).get(str(chat_id))
            if not rules:
                rules = Rules(str(chat_id))
            rules.rules = rules_text

            SESSION.add(rules)
            SESSION.commit()
        finally:
            SESSION.close()


def clear_rules(chat_id):
    with INSERTION_LOCK:
        try:
            rules = SESSION.query(Rules).get(str(chat_id))
            SESSION.delete(rules)
            SESSION.commit()
        except BaseException:
            return False
        finally:
            SESSION.close()
    return True


def get_rules(chat_id):
    with INSERTION_LOCK:
        try:
            rules = SESSION.query(Rules).get(str(chat_id))
            ret_rules = ""
            if rules:
                ret_rules = rules.rules
        except BaseException:
            return False
        finally:
            SESSION.close()
    return ret_rules


def num_chats():
    try:
        return SESSION.query(func.count(distinct(Rules.chat_id))).scalar()
    finally:
        SESSION.close()


def migrate_chat(old_chat_id, new_chat_id):
    with INSERTION_LOCK:
        try:
            chat = SESSION.query(Rules).get(str(old_chat_id))
            if chat:
                chat.chat_id = str(new_chat_id)
            SESSION.commit()
        finally:
            SESSION.close()
