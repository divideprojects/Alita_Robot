from time import time
from redis import Redis
from datetime import datetime
from os import path, mkdir, environ
from importlib import import_module as imp_mod
from sys import stdout, version_info, exit as sysexit
from logging import (
    FileHandler,
    StreamHandler,
    basicConfig,
    INFO,
    # WARNING,
    getLogger,
    DEBUG,
)

log_datetime = datetime.now().strftime("%d_%m_%Y-%H_%M_%S")
logdir = f"{__name__}/logs"

# Make Logs directory if it does not exixts
if not path.isdir(logdir):
    mkdir(logdir)

logfile = f"{logdir}/{__name__}_{log_datetime}.txt"

file_handler = FileHandler(filename=logfile)
stdout_handler = StreamHandler(stdout)

basicConfig(
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    level=INFO,
    handlers=[file_handler, stdout_handler],
)

getLogger("pyrogram").setLevel(DEBUG)
LOGGER = getLogger(__name__)

# if version < 3.6, stop bot.
if version_info[0] < 3 or version_info[1] < 7:
    LOGGER.error(
        (
            "You MUST have a Python Version of at least 3.7!\n"
            "Multiple features depend on this. Bot quitting."
        )
    )
    sysexit(1)  # Quit the Script

# the secret configuration specific things
try:
    if environ.get("ENV"):
        from alita.config import Config
    else:
        from alita.config import Development as Config
except Exception as ef:
    LOGGER.error(ef)  # Print Error
    sysexit(1)

# Redis Cache
redis_client = Redis(
    host=Config.REDIS_HOST, port=Config.REDIS_PORT, password=Config.REDIS_PASS
)

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
UPTIME = time()  # Check bot uptime
BOT_USERNAME = ""
BOT_NAME = ""


async def get_self(c):
    global BOT_USERNAME, BOT_NAME
    getbot = await c.get_me()
    BOT_NAME = getbot.first_name
    BOT_USERNAME = getbot.username
    return getbot


async def load_cmds(ALL_PLUGINS):
    for single in ALL_PLUGINS:
        imported_module = imp_mod("alita.plugins." + single)
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

    return ", ".join(list(HELP_COMMANDS.keys()))
