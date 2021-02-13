from pickle import loads, dumps
from alita import redisClient


async def set_key(key: str, value):
    redisClient.set(key, dumps(value))


async def get_key(key: str):
    return loads(redisClient.get(key))


async def flushredis():
    redisClient.flushall()
    return


async def allkeys():
    return redisClient.keys()
