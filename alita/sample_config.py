from os import environ


def load_var(var_name, def_value=None):
    return environ.get(var_name, def_value)


class Config:
    LOGGER = True
    TOKEN = load_var("TOKEN")
    APP_ID = int(load_var("APP_ID"))
    API_HASH = load_var("API_HASH")
    OWNER_ID = int(load_var("OWNER_ID"))
    MESSAGE_DUMP = int(load_var("MESSAGE_DUMP", -100))
    DEV_USERS = [int(i) for i in load_var("DEV_USERS", "").split()]
    SUDO_USERS = [int(i) for i in load_var("SUDO_USERS", "").split()]
    WHITELIST_USERS = [int(i) for i in load_var("WHITELIST_USERS", "").split()]
    DB_URI = load_var("DB_URI")
    REDIS_HOST = load_var("REDIS_HOST")
    REDIS_PORT = load_var("REDIS_PORT")
    REDIS_PASS = load_var("REDIS_PASS")
    NO_LOAD = load_var("NO_LOAD", "").split()
    PREFIX_HANDLER = load_var("PREFIX_HANDLER").split()
    SUPPORT_GROUP = load_var("SUPPORT_GROUP")
    SUPPORT_CHANNEL = load_var("SUPPORT_CHANNEL")
    ENABLED_LOCALES = [str(i) for i in load_var("ENABLED_LOCALES", "").split()]
    VERSION = load_var("VERSION")
    DEV_PREFIX_HANDLER = load_var("DEV_PREFIX_HANDLER", ">").split()
    WORKERS = int(load_var("WORKERS"))


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
    REDIS_PORT = "REDIS_PORT"  # int type
    REDIS_PASS = "REDIS_PASS"
    NO_LOAD = []
    PREFIX_HANDLER = ["!", "/"]
    SUPPORT_GROUP = "SUPPORT_GROUP"
    SUPPORT_CHANNEL = "SUPPORT_CHANNEL"
    ENABLED_LOCALES = "ENABLED_LOCALES"
    VERSION = "VERSION"
    DEV_PREFIX_HANDLER = ">"
    WORKERS = 8
