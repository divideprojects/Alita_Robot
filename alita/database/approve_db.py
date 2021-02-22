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

from sqlalchemy import Column, Integer, String

from alita.database import BASE, SESSION


class Approvals(BASE):
    __tablename__ = "approve"
    chat_id = Column(String(14), primary_key=True)
    user_id = Column(Integer, primary_key=True)

    def __init__(self, chat_id, user_id):
        self.chat_id = str(chat_id)  # ensure string
        self.user_id = user_id

    def __repr__(self):
        return f"<Approve {self.user_id}>"


Approvals.__table__.create(checkfirst=True)

INSERTION_LOCK = threading.RLock()


def approve(chat_id, user_id):
    with INSERTION_LOCK:
        note = Approvals(str(chat_id), user_id)
        SESSION.add(note)
        SESSION.commit()


def is_approved(chat_id, user_id):
    try:
        return SESSION.query(Approvals).get((str(chat_id), user_id))
    finally:
        SESSION.close()


def disapprove(chat_id, user_id):
    with INSERTION_LOCK:
        note = SESSION.query(Approvals).get((str(chat_id), user_id))
        if note:
            SESSION.delete(note)
            SESSION.commit()
            return True
        SESSION.close()
        return False


def all_approved(chat_id):
    try:
        return (
            SESSION.query(Approvals)
            .filter(Approvals.chat_id == str(chat_id))
            .order_by(Approvals.user_id.asc())
            .all()
        )
    finally:
        SESSION.close()


def disapprove_all(chat_id):
    users_list = []
    try:
        users = (
            SESSION.query(Approvals)
            .filter(Approvals.chat_id == str(chat_id))
            .order_by(Approvals.user_id.asc())
            .all()
        )
        for i in users:
            users_list.append(int(i.user_id))
        with INSERTION_LOCK:
            for user_id in users_list:
                note = SESSION.query(Approvals).get((str(chat_id), user_id))
                if note:
                    SESSION.delete(note)
                    SESSION.commit()
                else:
                    SESSION.close()
        return True
    except BaseException:
        return False
