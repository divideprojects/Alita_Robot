import threading
from sqlalchemy import Column, UnicodeText, Integer, String
from alita.db import BASE, SESSION

group_types = ("group", "supergroup")


class UserLang(BASE):
    __tablename__ = "language_user"

    user_id = Column(Integer, primary_key=True)
    lang_code = Column(UnicodeText)

    def __init__(self, user_id, lang_code="en-US"):
        self.user_id = user_id
        self.lang_code = lang_code

    def __repr__(self):
        return "Language for User {} is {}".format(self.user_id, self.user_lang)


class GroupLang(BASE):
    __tablename__ = "language_group"

    chat_id = Column(String, primary_key=True)
    lang_code = Column(UnicodeText)

    def __init__(self, chat_id, lang_code="en-US"):
        self.chat_id = chat_id
        self.lang_code = lang_code

    def __repr__(self):
        return "Language for Group {} is {}".format(self.user_id, self.group_lang)


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
    default_lang = "en-US"
    with INSERTION_LOCK:
        if chat_type == "private":
            try:
                exist = SESSION.query(UserLang).get(chat_id)
                if exist:
                    lang = exist.lang_code
                    return lang
                exist = UserLang(chat_id, default_lang)
                return default_lang
            finally:
                SESSION.close()
        elif chat_type in group_types:
            try:
                exist = SESSION.query(GroupLang).get(str(chat_id))
                if exist:
                    lang = exist.lang_code
                    return lang
                exist = GroupLang(str(chat_id), default_lang)
                return default_lang
            finally:
                SESSION.close()


def migrate_chat(old_chat_id, new_chat_id):
    with INSERTION_LOCK:
        chat = SESSION.query(GroupLang).get(str(old_chat_id))
        if chat:
            chat.chat_id = str(new_chat_id)
            SESSION.merge(chat)
        SESSION.commit()
        SESSION.close()
