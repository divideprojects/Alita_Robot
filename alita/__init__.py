import sys
import os
import time
import pickle
import logging
import importlib
import redis
from pyrogram import Client

logging.basicConfig(
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    level=logging.INFO,
)

LOGGER = logging.getLogger(__name__)

# if version < 3.6, stop bot.
if sys.version_info[0] < 3 or sys.version_info[1] < 6:
    LOGGER.error(
        (
            "You MUST have a Python Version of at least 3.6!\n"
            "Multiple features depend on this. Bot quitting."
        )
    )
    quit(1)  # Quit the Script

# the secret configuration specific things
try:
    if bool(os.environ.get("ENV", False)):
        from alita.sample_config import Config
    else:
        from alita.config import Development as Config
except Exception as ef:
    print(ef)  # Print Error

# Redis Cache
REDIS_HOST = Config.REDIS_HOST
REDIS_PORT = Config.REDIS_PORT
REDIS_DB = Config.REDIS_DB
redisClient = redis.Redis(host=REDIS_HOST, port=REDIS_PORT, db=REDIS_DB)

# Account Related
TOKEN = Config.TOKEN
APP_ID = Config.APP_ID
API_HASH = Config.API_HASH

# General Config
MESSAGE_DUMP = Config.MESSAGE_DUMP
SUPPORT_GROUP = Config.SUPPORT_GROUP
SUPPORT_CHANNEL = Config.SUPPORT_CHANNEL

# Users Config
OWNER_ID = Config.OWNER_ID
DEV_USERS = Config.DEV_USERS
SUDO_USERS = Config.SUDO_USERS
WHITELIST_USERS = Config.WHITELIST_USERS
SUPPORT_STAFF = list(
    dict.fromkeys([OWNER_ID] + SUDO_USERS + DEV_USERS + WHITELIST_USERS)
)  # Remove duplicates!

# Plugins, DB and Workers
DB_URI = Config.DB_URI
NO_LOAD = Config.NO_LOAD
WORKERS = Config.WORKERS

# Prefixes
PREFIX_HANDLER = Config.PREFIX_HANDLER
DEV_PREFIX_HANDLER = Config.DEV_PREFIX_HANDLER
ENABLED_LOCALES = Config.ENABLED_LOCALES
VERSION = Config.VERSION

HELP_COMMANDS = {}  # For help menu
UPTIME = time.time()  # Check bot uptime


def load_staff():
    begin = time.time()
    LOGGER.info("Begin Caching Support Staff...")
    LOGGER.info(SUPPORT_STAFF)
    try:
        redisClient.set(
            "SUPPORT_STAFF", pickle.dumps(SUPPORT_STAFF)
        )  # Redis set value of support staff
        LOGGER.info(
            f"Cached SUPPORT_STAFF\nTime Taken: {round(time.time() - begin, 5)}s"
        )
    except Exception as ef:
        LOGGER.info(ef)


def load_cmds(ALL_PLUGINS):
    for single in ALL_PLUGINS:
        imported_module = importlib.import_module("alita.plugins." + single)
        if not hasattr(imported_module, "__PLUGIN__"):
            imported_module.__PLUGIN__ = imported_module.__name__

        if not imported_module.__PLUGIN__.lower() in HELP_COMMANDS:
            if hasattr(imported_module, "__help__") and imported_module.__help__:
                HELP_COMMANDS[
                    imported_module.__PLUGIN__.lower()
                ] = imported_module.__help__
            else:
                continue
        else:
            raise Exception(
                "Can't have two plugins with the same name! Please change one"
            )

    return f"Plugins Loaded: {list(list(HELP_COMMANDS.keys()))}"
