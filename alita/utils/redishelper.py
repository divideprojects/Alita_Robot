from pickle import loads, dumps
from alita import redis_client


async def set_key(key: str, value):
    return redis_client.set(key, dumps(value))


async def get_key(key: str):
    return loads(redis_client.get(key))


async def flushredis():
    return redis_client.flushall()


async def allkeys():
    keys = redis_client.keys()
    keys_str = []
    for i in keys:
        keys_str.append(i.decode())
    return keys_str
