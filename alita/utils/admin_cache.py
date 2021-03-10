from time import perf_counter

from cachetools import TTLCache

# admins stay cached for one hour
ADMIN_CACHE = TTLCache(maxsize=512, ttl=(60 * 60), timer=perf_counter)
