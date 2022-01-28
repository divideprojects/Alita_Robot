from functools import reduce
from glob import glob
from operator import getitem
from os import path
from threading import RLock
from traceback import format_exc

from pyrogram.types import CallbackQuery
from yaml import FullLoader
from yaml import load as load_yml

from alita import ENABLED_LOCALES, LOGGER
from alita.database.lang_db import Langs

# Initialise
LANG_LOCK = RLock()


def cache_localizations(files):
    """Get all translated strings from files."""
    ldict = {lang: {} for lang in ENABLED_LOCALES}
    for file in files:
        lang_name = (file.split(path.sep)[1]).replace(".yml", "")
        lang_data = load_yml(open(file, encoding="utf-8"), Loader=FullLoader)
        ldict[lang_name] = lang_data
    return ldict


# Get all translation files
lang_files = []
for locale in ENABLED_LOCALES:
    lang_files += glob(path.join("locales", f"{locale}.yml"))
lang_dict = cache_localizations(lang_files)


def tlang(m, user_msg):
    """Main function for getting the string of preferred language."""
    with LANG_LOCK:
        default_lang = "en"

        m_args = user_msg.split(".")  # Split in a list

        # Get Chat
        if isinstance(m, CallbackQuery):
            m = m.message

        # Get language of user from database, default = 'en' (English)
        try:
            lang = Langs(m.chat.id).get_lang()
        except Exception as ef:
            LOGGER.error(f"Lang Error: {ef}")
            lang = default_lang
            LOGGER.error(format_exc())

        # Raise exception if lang_code not found
        if lang not in ENABLED_LOCALES:
            LOGGER.error("Non-enabled locale used by user!")
            lang = default_lang

        # Get lang
        m_args.insert(0, lang)
        m_args.insert(1, "strings")

        try:
            txt = reduce(getitem, m_args, lang_dict)
        except KeyError:
            m_args.pop(0)
            m_args.insert(0, default_lang)
            txt = reduce(getitem, m_args, lang_dict)
            LOGGER.error(format_exc())

        return txt
