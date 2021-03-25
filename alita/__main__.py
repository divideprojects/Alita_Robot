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


from time import time

from alita import LOGGER
from alita.bot_class import Alita
from alita.database.antichannelpin_db import __load_pins_chats
from alita.database.antispam_db import __load_antispam_users
from alita.database.approve_db import __load_approve_cache
from alita.database.chats_db import __load_chats_cache
from alita.database.filters_db import __load_filters_cache
from alita.database.group_blacklist import __load_group_blacklist
from alita.database.lang_db import __load_all_langs
from alita.database.reporting_db import __load_all_reporting_settings
from alita.database.rules_db import __load_all_rules
from alita.database.users_db import __load_users_cache


def load_caches():
    # Load local cache dictionaries
    start = time()
    LOGGER.info("Starting to load Local Caches!")
    __load_all_langs()
    __load_chats_cache()
    __load_antispam_users()
    __load_users_cache()
    __load_filters_cache()
    __load_all_rules()
    __load_approve_cache()
    __load_pins_chats()
    __load_all_reporting_settings()
    __load_group_blacklist()
    LOGGER.info(f"Succefully loaded Local Caches in {round((time()-start),3)}s\n")


if __name__ == "__main__":
    load_caches()
    Alita().run()
