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


from traceback import format_exc
from typing import Tuple

from alita import LOGGER
from alita.database.users_db import Users

# Initialize
db = Users()


async def extract_user(c, m) -> Tuple[int, str, str]:
    """Extract the user from the provided message."""
    user_id = None
    user_first_name = None
    user_name = None

    if m.reply_to_message and m.reply_to_message.from_user:
        user_id = m.reply_to_message.from_user.id
        user_first_name = m.reply_to_message.from_user.first_name
        user_name = m.reply_to_message.from_user.username

    elif len(m.text.split()) > 1:
        if len(m.entities) > 1:
            required_entity = m.entities[1]
            if required_entity.type == "text_mention":
                user_id = required_entity.user.id
                user_first_name = required_entity.user.first_name
                user_name = required_entity.user.username
            elif required_entity.type in ("mention", "phone_number"):
                # new long user ids are identified as phone_number
                user_found = m.text[
                    required_entity.offset : (
                        required_entity.offset + required_entity.length
                    )
                ]

                try:
                    user_found = int(user_found)
                except (ValueError, Exception) as ef:
                    if "invalid literal for int() with base 10:" in str(ef):
                        user_found = str(user_found)
                    else:
                        user_found = user_found
                        LOGGER.error(ef)
                        LOGGER.error(format_exc())

                try:
                    user = db.get_user_info(user_found)
                    user_id = user["_id"]
                    user_first_name = user["name"]
                    user_name = user["username"]
                except KeyError:
                    # If user not in database
                    user = await c.get_users(user_found)
                    user_id = user.id
                    user_first_name = user.first_name
                    user_name = user.username
                except Exception as ef:
                    user_id = user_found
                    user_first_name = user_found
                    user_name = ""
                    LOGGER.error(ef)
                    LOGGER.error(format_exc())

        else:
            try:
                user_id = int(m.text.split()[1])
            except (ValueError, Exception) as ef:
                if "invalid literal for int() with base 10:" in str(ef):
                    user_id = str(m.text.split()[1])
                else:
                    user_id = m.text.split()[1]
                    LOGGER.error(ef)
                    LOGGER.error(format_exc())
            try:
                user = db.get_user_info(user_id)
                user_first_name = user["name"]
                user_name = user["username"]
            except Exception as ef:
                user_id = m.text.split()[1]
                user = await c.get_users(user_id)
                user_first_name = user.first_name
                user_name = user.username
                LOGGER.error(ef)
                LOGGER.error(format_exc())

    else:
        user_id = m.from_user.id
        user_first_name = m.from_user.first_name
        user_name = m.from_user.username

    return user_id, user_first_name, user_name
