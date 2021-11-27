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
from httpx import AsyncClient
from httpx import Timeout

timeout = Timeout(40, pool=None)
http = AsyncClient(http2=True, timeout=timeout)


class HTTPx:
    """class for helping get the data from url using aiohttp."""

    @staticmethod
    async def get(link: str):
        """Get JSON data from the provided link."""
        async with AsyncClient() as sess:
            return await sess.get(link)
