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


async def extract_user(c, m) -> Tuple[int, str]:
    """Extract the user from the provided message."""
    user_id = None
    user_first_name = None

    if m.reply_to_message:
        user_id = m.reply_to_message.from_user.id
        user_first_name = m.reply_to_message.from_user.first_name

    elif len(m.command) > 1:
        if len(m.entities) > 1:
            required_entity = m.entities[1]
            if required_entity.type == "text_mention":
                user_id = required_entity.user.id
                user_first_name = required_entity.user.first_name
            elif required_entity.type == "mention":
                user_found = m.text[
                    required_entity.offset : (
                        required_entity.offset + required_entity.length
                    )
                ]
                try:
                    user = db.get_user_info(user_found)
                    user_id = user["_id"]
                    user_first_name = user["name"]
                except KeyError:
                    user = await c.get_users(user_found)
                    user_id = user.id
                    user_first_name = user.first_name
                except Exception as ef:
                    user_id = user_found
                    user_first_name = user_found
                    LOGGER.error(ef)
                    LOGGER.error(format_exc())

            # new long user_ids are identified as phone_number
            elif required_entity.type == "phone_number":
                user_found = m.text[
                    required_entity.offset : (
                        required_entity.offset + required_entity.length
                    )
                ]
                try:
                    user = db.get_user_info(user_found)
                    user_id = user["_id"]
                    user_first_name = user["name"]
                except KeyError:
                    user = await c.get_users(user_found)
                    user_id = user.id
                    user_first_name = user.first_name
                except Exception as ef:
                    user_id = user_found
                    user_first_name = user_found
                    LOGGER.error(ef)
                    LOGGER.error(format_exc())
        else:
            user_id = int(m.command[1])
            try:
                user_first_name = db.get_user_info(int(user_id))["name"]
            except Exception as ef:
                user_first_name = (await c.get_users(int(user_id))).first_name
                LOGGER.error(ef)
                LOGGER.error(format_exc())

    else:
        user_id = m.from_user.id
        user_first_name = m.from_user.first_name

    return user_id, user_first_name
