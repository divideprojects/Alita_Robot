from httpx import AsyncClient, Timeout

timeout = Timeout(40, pool=None)
http = AsyncClient(http2=True, timeout=timeout)


class HTTPx:
    """class for helping get the data from url using aiohttp."""

    @staticmethod
    async def get(link: str):
        """Get JSON data from the provided link."""
        async with AsyncClient() as sess:
            return await sess.get(link)
