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


from threading import RLock
from time import perf_counter

from cachetools import TTLCache

from alita import LOGGER

THREAD_LOCK = RLock()

# admins stay cached for one hour
ADMIN_CACHE = TTLCache(maxsize=512, ttl=(60 * 60), timer=perf_counter)


async def admin_cache_reload(m):
    with THREAD_LOCK:
        global ADMIN_CACHE
        try:
            admin_list = [user[0] for user in ADMIN_CACHE[str(m.chat.id)]]
            return admin_list
        except KeyError:
            LOGGER.info(f"Loading admins for chat {m.chat.id}")
            admin_list = []
            async for i in m.chat.iter_members(filter="administrators"):
                admin_list.append(
                    (
                        i.user.id,
                        ("@" + i.user.username)
                        if i.user.username
                        else i.user.first_name,
                    ),
                )
            ADMIN_CACHE[str(m.chat.id)] = admin_list
            admin_list = [user[0] for user in admin_list]

        return False
