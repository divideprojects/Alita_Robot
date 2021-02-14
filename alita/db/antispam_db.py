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

from sqlalchemy import Column, Integer, UnicodeText

from alita.db import BASE, SESSION


class GloballyBannedUsers(BASE):
    __tablename__ = "gbans"
    user_id = Column(Integer, primary_key=True)
    name = Column(UnicodeText, nullable=False)
    reason = Column(UnicodeText)

    def __init__(self, user_id, name, reason=None):
        self.user_id = user_id
        self.name = name
        self.reason = reason

    def __repr__(self):
        return f"<GBanned User {self.name} ({self.user_id})>"

    def to_dict(self):
        return {"user_id": self.user_id, "name": self.name, "reason": self.reason}


GloballyBannedUsers.__table__.create(checkfirst=True)
GBAN_LOCK = threading.RLock()

GBANNED_DICT = {}


def gban_user(user_id, name, reason=None):
    with GBAN_LOCK:
        user = SESSION.query(GloballyBannedUsers).get(user_id)
        try:
            if not user:
                user = GloballyBannedUsers(user_id, name, reason)
            else:
                user.reason = reason
                user.name = name
            SESSION.merge(user)
            SESSION.commit()
        finally:
            SESSION.close()
    __load_gbanned_userid_list()


def update_gban_reason(user_id, name, reason=None):
    with GBAN_LOCK:
        try:
            user = SESSION.query(GloballyBannedUsers).get(user_id)
            if not user:
                user = GloballyBannedUsers(user_id, name, reason)
                old_reason = ""
            else:
                old_reason = user.reason
                user.name = name
                user.reason = reason
            SESSION.merge(user)
            SESSION.commit()
        finally:
            SESSION.close()
    return old_reason


def ungban_user(user_id):
    with GBAN_LOCK:
        user = SESSION.query(GloballyBannedUsers).get(user_id)
        try:
            if user:
                SESSION.delete(user)
                SESSION.commit()
        finally:
            SESSION.close()
            __load_gbanned_userid_list()


def is_user_gbanned(user_id):
    return user_id in list(GBANNED_DICT.keys())


def get_gban_list():
    try:
        return [x.to_dict() for x in SESSION.query(GloballyBannedUsers).all()]
    finally:
        SESSION.close()


def num_gbanned_users():
    return len(list(GBANNED_DICT.keys()))


def __load_gbanned_userid_list():
    global GBANNED_DICT
    getall = SESSION.query(GloballyBannedUsers).all()
    try:
        for x in getall:
            GBANNED_DICT[x.user_id] = {"name": x.name, "reason": x.reason}
    finally:
        SESSION.close()


# Create in memory to avoid over disk usage for SQL
__load_gbanned_userid_list()
