import pickle
from alita import redisClient


async def set_key(key: str, value):
    redisClient.set(key, pickle.dumps(value))


async def get_key(key: str):
    return pickle.loads(redisClient.get(key))


async def flushredis():
    redisClient.flushall()
    return


async def allkeys():
    return redisClient.keys()
