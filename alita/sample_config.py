import os


class Config:
    LOGGER = True
    TOKEN = os.environ.get("TOKEN")
    APP_ID = int(os.environ.get("APP_ID"))
    API_HASH = os.environ.get("API_HASH")
    OWNER_ID = int(os.environ.get("OWNER_ID"))
    MESSAGE_DUMP = int(os.environ.get("MESSAGE_DUMP"))
    DEV_USERS = os.environ.get("DEV_USERS").split()
    SUDO_USERS = os.environ.get("SUDO_USERS").split()
    WHITELIST_USERS = os.environ.get("WHITELIST_USERS").split()
    DB_URI = os.environ.get("DB_URI")
    REDIS_HOST = os.environ.get("REDIS_HOST")
    REDIS_PORT = os.environ.get("REDIS_PORT")
    REDIS_DB = os.environ.get("REDIS_DB")
    NO_LOAD = os.environ.get("NO_LOAD").split()
    PREFIX_HANDLER = os.environ.get("PREFIX_HANDLER").split()
    SUPPORT_GROUP = os.environ.get("SUPPORT_GROUP")
    SUPPORT_CHANNEL = os.environ.get("SUPPORT_CHANNEL")
    ENABLED_LOCALES = os.environ.get("ENABLED_LOCALES")
    VERSION = os.environ.get("VERSION")
    DEV_PREFIX_HANDLER = os.environ.get("DEV_PREFIX_HANDLER").split()
    WORKERS = int(os.environ.get("WORKERS"))


class Development:
    # Fill in these vars if you want to use Traditional methods
    LOGGER = True
    TOKEN = "YOUR TOKEN"
    APP_ID = 12345  # Your APP_ID - int value
    API_HASH = "YOUR TOKEN"
    OWNER_ID = "YOUR TOKEN"
    MESSAGE_DUMP = "YOUR TOKEN"
    DEV_USERS = []
    SUDO_USERS = []
    WHITELIST_USERS = []
    DB_URI = "postgres://username:password@postgresdb:5432/database_name"
    REDIS_HOST = "REDIS_HOST"
    REDIS_PORT = "REDIS_PORT"
    REDIS_DB = "REDIS_DB"
    NO_LOAD = []
    PREFIX_HANDLER = ["!", "/"]
    SUPPORT_GROUP = "SUPPORT_GROUP"
    SUPPORT_CHANNEL = "SUPPORT_CHANNEL"
    ENABLED_LOCALES = "ENABLED_LOCALES"
    VERSION = "VERSION"
    DEV_PREFIX_HANDLER = ">"
    WORKERS = 8
