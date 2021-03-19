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


from aiohttp import ClientSession


async def paste(content):
    """Paste the provided content to nekobin."""
    content = str(content)
    NEKOBIN_URL = "https://nekobin.com/"
    async with ClientSession() as sess:
        async with sess.post(
            NEKOBIN_URL + "api/documents",
            json={"content": content},
        ) as resp:
            if resp.status == 201:
                response = await resp.json()
                key = response["result"]["key"]
                final_url = f"{NEKOBIN_URL}{key}.txt"
                raw = f"{NEKOBIN_URL}raw/{key}.txt"
            else:
                raise Exception("Error Pasting to Nekobin")

    return final_url, raw
