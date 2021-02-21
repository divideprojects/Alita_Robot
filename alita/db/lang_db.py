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

from sqlalchemy import Column, Integer, String, UnicodeText

from alita.db import BASE, SESSION

group_types = ("group", "supergroup")


class UserLang(BASE):
    __tablename__ = "language_user"

    user_id = Column(Integer, primary_key=True)
    lang_code = Column(UnicodeText)

    def __init__(self, user_id, lang_code="en"):
        self.user_id = user_id
        self.lang_code = lang_code

    def __repr__(self):
        return f"Language for User {self.user_id} is {self.user_lang}"


class GroupLang(BASE):
    __tablename__ = "language_group"

    chat_id = Column(String, primary_key=True)
    lang_code = Column(UnicodeText)

    def __init__(self, chat_id, lang_code="en"):
        self.chat_id = chat_id
        self.lang_code = lang_code

    def __repr__(self):
        return f"Language for Group {self.user_id} is {self.group_lang}"


GroupLang.__table__.create(checkfirst=True)
UserLang.__table__.create(checkfirst=True)

INSERTION_LOCK = threading.RLock()


def set_lang(chat_id, chat_type, lang_code):
    with INSERTION_LOCK:
        if chat_type == "private":
            try:
                lang = SESSION.query(UserLang).get(chat_id)
                if not lang:
                    lang = UserLang(chat_id, lang_code)
                else:
                    lang.lang_code = lang_code
                SESSION.merge(lang)
                SESSION.commit()
            finally:
                SESSION.close()
        elif chat_type in group_types:
            try:
                lang = SESSION.query(GroupLang).get(str(chat_id))
                if not lang:
                    lang = GroupLang(str(chat_id), lang_code)
                else:
                    lang.lang_code = lang_code
                SESSION.merge(lang)
                SESSION.commit()
            finally:
                SESSION.close()


def get_lang(chat_id, chat_type):
    default_lang = "en"
    with INSERTION_LOCK:
        if chat_type == "private":
            try:
                exist = SESSION.query(UserLang).get(chat_id)
                if exist:
                    lang = exist.lang_code
                else:
                    exist = UserLang(chat_id, default_lang)
                    lang = default_lang
            finally:
                SESSION.close()
        elif chat_type in group_types:
            try:
                exist = SESSION.query(GroupLang).get(str(chat_id))
                if exist:
                    lang = exist.lang_code
                else:
                    exist = GroupLang(str(chat_id), default_lang)
                    lang = default_lang
            finally:
                SESSION.close()
    return lang


def migrate_chat(old_chat_id, new_chat_id):
    with INSERTION_LOCK:
        chat = SESSION.query(GroupLang).get(str(old_chat_id))
        if chat:
            chat.chat_id = str(new_chat_id)
            SESSION.merge(chat)
        SESSION.commit()
        SESSION.close()
