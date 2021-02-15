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


from pickle import dumps, loads


async def set_key(key: str, value):
    from alita import redis_client

    return await redis_client.set(key, dumps(value))


async def get_key(key: str):
    from alita import redis_client

    return loads(await redis_client.get(key))


async def flushredis():
    from alita import redis_client

    return await redis_client.flushall()


async def allkeys():
    from alita import redis_client

    keys = await redis_client.keys(pattern="*")
    keys_str = []
    for i in keys:
        keys_str.append(i.decode())
    return keys_str


async def close():
    from alita import redis_client

    redis_client.close()
    return await redis_client.wait_closed()
