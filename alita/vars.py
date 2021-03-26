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


from os import environ


def load_var(var_name, def_value=None):
    return environ.get(var_name, def_value)


class Config:
    """Config class for variables."""

    LOGGER = True
    STRING_SESSION = load_var("STRING_SESSION")
    APP_ID = int(load_var("APP_ID"))
    API_HASH = load_var("API_HASH")
    OWNER_ID = int(load_var("OWNER_ID"))
    MESSAGE_DUMP = int(load_var("MESSAGE_DUMP", -100))
    DEV_USERS = [int(i) for i in load_var("DEV_USERS", "").split()]
    SUDO_USERS = [int(i) for i in load_var("SUDO_USERS", "").split()]
    WHITELIST_USERS = [int(i) for i in load_var("WHITELIST_USERS", "").split()]
    DB_URI = load_var("DB_URI")
    DB_NAME = load_var("DB_NAME", "alita_robot")
    NO_LOAD = load_var("NO_LOAD", "").split()
    PREFIX_HANDLER = load_var("PREFIX_HANDLER", "/").split()
    SUPPORT_GROUP = load_var("SUPPORT_GROUP")
    SUPPORT_CHANNEL = load_var("SUPPORT_CHANNEL")
    ENABLED_LOCALES = [str(i) for i in load_var("ENABLED_LOCALES", "").split()]
    VERSION = load_var("VERSION")
    DEV_PREFIX_HANDLER = load_var("DEV_PREFIX_HANDLER", ">").split()
    WORKERS = int(load_var("WORKERS", 16))


class Development:
    """Development class for variables."""

    # Fill in these vars if you want to use Traditional method of deploying
    LOGGER = True
    STRING_SESSION = "YOUR STRING_SESSION"
    APP_ID = 12345  # Your APP_ID from Telegram
    API_HASH = "YOUR TOKEN"  # Your APP_HASH from Telegram
    OWNER_ID = 12345  # Your telegram user id
    MESSAGE_DUMP = "YOUR GROUP_ID"  # Your Private Group ID
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
    DEV_PREFIX_HANDLER = "?"
    WORKERS = 8
