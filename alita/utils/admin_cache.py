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
from time import perf_counter, time

from cachetools import TTLCache
from pyrogram.types import CallbackQuery

from alita import LOGGER

THREAD_LOCK = RLock()

# admins stay cached for 30 mins
ADMIN_CACHE = TTLCache(maxsize=512, ttl=(60 * 30), timer=perf_counter)
# Block from refreshing admin list for 10 mins
TEMP_ADMIN_CACHE_BLOCK = TTLCache(maxsize=512, ttl=(60 * 10), timer=perf_counter)


async def admin_cache_reload(m, status=None):
    start = time()
    with THREAD_LOCK:

        if isinstance(m, CallbackQuery):
            m = m.message

        global ADMIN_CACHE, TEMP_ADMIN_CACHE_BLOCK
        if status is not None:
            TEMP_ADMIN_CACHE_BLOCK[m.chat.id] = status

        try:
            if TEMP_ADMIN_CACHE_BLOCK[m.chat.id] in ("autoblock", "manualblock"):
                return
        except KeyError:
            # Because it might be first time when admn_list is being reloaded
            pass

        admin_list = [
            (
                z.user.id,
                (("@" + z.user.username) if z.user.username else z.user.first_name),
                z.is_anonymous,
            )
            async for z in m.chat.iter_members(filter="administrators")
            if not z.user.is_deleted
        ]
        ADMIN_CACHE[m.chat.id] = admin_list
        LOGGER.info(
            f"Loaded admins for chat {m.chat.id} in {round((time()-start),3)}s due to '{status}'",
        )
        TEMP_ADMIN_CACHE_BLOCK[m.chat.id] = "autoblock"

        return admin_list
