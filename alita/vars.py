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
from os import getcwd

from prettyconf import Configuration
from prettyconf.loaders import EnvFile
from prettyconf.loaders import Environment

env_file = f"{getcwd()}/.env"
config = Configuration(loaders=[Environment(), EnvFile(filename=env_file)])


class Config:
    """Config class for variables."""

    LOGGER = True
    BOT_TOKEN = config("BOT_TOKEN", default=None)
    APP_ID = int(config("APP_ID", default=None))
    API_HASH = config("API_HASH", default=None)
    OWNER_ID = int(config("OWNER_ID", default=1198820588))
    MESSAGE_DUMP = int(config("MESSAGE_DUMP", default=-100))
    DEV_USERS = [int(i) for i in config("DEV_USERS", default="").split()]
    SUDO_USERS = [int(i) for i in config("SUDO_USERS", default="").split()]
    WHITELIST_USERS = [
        int(i) for i in config("WHITELIST_USERS", default="").split()
    ]
    DB_URI = config("DB_URI", default="")
    DB_NAME = config("DB_NAME", default="alita_robot")
    NO_LOAD = config("NO_LOAD", default="").split()
    PREFIX_HANDLER = config("PREFIX_HANDLER", default="/").split()
    SUPPORT_GROUP = config("SUPPORT_GROUP", default="DivideProjectsDiscussion")
    SUPPORT_CHANNEL = config("SUPPORT_CHANNEL", default="DivideProjects")
    ENABLED_LOCALES = [
        str(i) for i in config("ENABLED_LOCALES", default="en").split()
    ]
    VERSION = config("VERSION", default="v2.0")
    WORKERS = int(config("WORKERS", default=16))
    BOT_USERNAME = ""
    BOT_ID = ""
    BOT_NAME = ""


class Development:
    """Development class for variables."""

    # Fill in these vars if you want to use Traditional method of deploying
    LOGGER = True
    BOT_TOKEN = "YOUR BOT_TOKEN"
    APP_ID = 12345  # Your APP_ID from Telegram
    API_HASH = "YOUR API HASH"  # Your APP_HASH from Telegram
    OWNER_ID = 12345  # Your telegram user id
    MESSAGE_DUMP = -100  # Your Private Group ID for logs
    DEV_USERS = []
    SUDO_USERS = []
    WHITELIST_USERS = []
    DB_URI = "postgres://username:password@postgresdb:5432/database_name"
    DB_NAME = "alita_robot"
    NO_LOAD = []
    PREFIX_HANDLER = ["!", "/"]
    SUPPORT_GROUP = "SUPPORT_GROUP"
    SUPPORT_CHANNEL = "SUPPORT_CHANNEL"
    ENABLED_LOCALES = ["ENABLED_LOCALES"]
    VERSION = "VERSION"
    WORKERS = 8
