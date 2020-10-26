import threading
from alita.db import BASE, SESSION
from sqlalchemy import Column, String, UnicodeText, func, distinct


class Rules(BASE):
    __tablename__ = "rules"
    chat_id = Column(String(14), primary_key=True)
    rules = Column(UnicodeText, default="")

    def __init__(self, chat_id):
        self.chat_id = chat_id

    def __repr__(self):
        return "<Chat {} rules: {}>".format(self.chat_id, self.rules)


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
