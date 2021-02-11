from aiohttp import ClientSession


class AioHttp:
    """class for helping get the data from url using aiohttp."""

    @staticmethod
    async def get_json(link):
        async with ClientSession() as session:
            async with session.get(link) as resp:
                return await resp.json(), resp

    @staticmethod
    async def get_text(link):
        async with ClientSession() as session:
            async with session.get(link) as resp:
                return await resp.text(), resp

    @staticmethod
    async def get_raw(link):
        async with ClientSession() as session:
            async with session.get(link) as resp:
                return await resp.read(), resp

    @staticmethod
    async def post(link):
        async with ClientSession() as session:
            async with session.post(link) as resp:
                return resp
