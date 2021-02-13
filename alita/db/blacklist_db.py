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
from sqlalchemy import func, distinct, Column, String, UnicodeText
from alita.db import BASE, SESSION


class BlackListFilters(BASE):
    __tablename__ = "blacklist"
    chat_id = Column(String(14), primary_key=True)
    trigger = Column(UnicodeText, primary_key=True, nullable=False)

    def __init__(self, chat_id, trigger):
        self.chat_id = str(chat_id)
        self.trigger = trigger

    def __repr__(self):
        return f"<Blacklist filter '{self.trigger}' for {self.chat_id}>"

    def __eq__(self, other):
        return bool(
            isinstance(other, BlackListFilters)
            and self.chat_id == other.chat_id
            and self.trigger == other.trigger
        )


BlackListFilters.__table__.create(checkfirst=True)
INSERTION_LOCK = threading.RLock()
CHAT_BLACKLISTS = {}


def add_to_blacklist(chat_id, trigger):
    with INSERTION_LOCK:
        try:
            blacklist_filt = BlackListFilters(str(chat_id), trigger)

            SESSION.merge(blacklist_filt)
            SESSION.commit()
            CHAT_BLACKLISTS.setdefault(str(chat_id), set()).add(trigger)
        finally:
            SESSION.close()


def rm_from_blacklist(chat_id, trigger):
    with INSERTION_LOCK:
        try:
            blacklist_filt = SESSION.query(BlackListFilters).get(
                (str(chat_id), trigger)
            )
            if blacklist_filt:
                if trigger in CHAT_BLACKLISTS.get(str(chat_id), set()):
                    CHAT_BLACKLISTS.get(str(chat_id), set()).remove(trigger)

                SESSION.delete(blacklist_filt)
                SESSION.commit()
                return True
        finally:
            SESSION.close()
        return False


def get_chat_blacklist(chat_id):
    if CHAT_BLACKLISTS.get(str(chat_id), set()):
        return CHAT_BLACKLISTS.get(str(chat_id), set())
    return False


def num_blacklist_filters():
    try:
        return SESSION.query(BlackListFilters).count()
    finally:
        SESSION.close()


def num_blacklist_chat_filters(chat_id):
    try:
        return (
            SESSION.query(BlackListFilters.chat_id)
            .filter(BlackListFilters.chat_id == str(chat_id))
            .count()
        )
    finally:
        SESSION.close()


def num_blacklist_filter_chats():
    try:
        return SESSION.query(func.count(distinct(BlackListFilters.chat_id))).scalar()
    finally:
        SESSION.close()


def __load_chat_blacklists():
    global CHAT_BLACKLISTS
    try:
        chats = SESSION.query(BlackListFilters.chat_id).distinct().all()
        for (chat_id,) in chats:
            CHAT_BLACKLISTS[chat_id] = []

        all_filters = SESSION.query(BlackListFilters).all()
        for x in all_filters:
            CHAT_BLACKLISTS[x.chat_id] += [x.trigger]

        CHAT_BLACKLISTS = {x: set(y) for x, y in CHAT_BLACKLISTS.items()}
    finally:
        SESSION.close()


def migrate_chat(old_chat_id, new_chat_id):
    with INSERTION_LOCK:
        try:
            chat_filters = (
                SESSION.query(BlackListFilters)
                .filter(BlackListFilters.chat_id == str(old_chat_id))
                .all()
            )
            for filt in chat_filters:
                filt.chat_id = str(new_chat_id)
            SESSION.commit()
        finally:
            SESSION.close()


__load_chat_blacklists()
