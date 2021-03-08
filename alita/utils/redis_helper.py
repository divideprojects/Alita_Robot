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


from aioredis import create_redis_pool
from ujson import dumps, loads

from alita import LOGGER, REDIS_DB, REDIS_HOST, REDIS_PASS, REDIS_PORT

# Initialize redis_client var
redis_client = None


async def setup_redis():
    """Start redis client."""
    global redis_client
    redis_client = await create_redis_pool(
        address=(REDIS_HOST, REDIS_PORT),
        db=REDIS_DB,
        password=REDIS_PASS,
    )
    try:
        await redis_client.ping()
        return redis_client
    except Exception as ef:
        LOGGER.error(f"Cannot connect to redis\nError: {ef}")
        return False


class RedisHelper:
    """Class for connecting to Redis cache db."""

    @staticmethod
    async def set_key(key: str, value):
        """Set the key data in Redis Cache."""
        return await redis_client.set(
            key,
            dumps(
                value,
                reject_bytes=False,
                escape_forward_slashes=True,
                encode_html_chars=True,
            ),
        )

    @staticmethod
    async def get_key(key: str):
        """Get the key data from Redis Cache."""
        return loads(await redis_client.get(key))

    @staticmethod
    async def flushredis():
        """Empty the Redis Cache Database."""
        return await redis_client.flushall()

    @staticmethod
    async def allkeys():
        """Get all keys from Redis Cache."""
        keys = await redis_client.keys(pattern="*")
        keys_str = []
        for i in keys:
            keys_str.append(i.decode())
        return keys_str

    @staticmethod
    async def close():
        """Close connection to Redis."""
        redis_client.close()
        return await redis_client.wait_closed()
