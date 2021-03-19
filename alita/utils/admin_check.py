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

from pyrogram.types import CallbackQuery, Message

from alita import DEV_USERS, LOGGER, OWNER_ID, SUDO_USERS

SUDO_LEVEL = SUDO_USERS + DEV_USERS + [int(OWNER_ID)]
DEV_LEVEL = DEV_USERS + [int(OWNER_ID)]


async def admin_check(m) -> bool:
    """Checks if user is admin or not."""
    if isinstance(m, Message):
        user_id = m.from_user.id
    if isinstance(m, CallbackQuery):
        user_id = m.message.from_user.id

    try:
        if user_id in SUDO_LEVEL:
            return True
    except Exception as ef:
        LOGGER.error(format_exc())

    user = await m.chat.get_member(user_id)
    admin_strings = ("creator", "administrator")

    if user.status not in admin_strings:
        reply = "Nigga, you're not admin, don't try this explosive shit."
        try:
            await m.edit_text(reply)
        except Exception as ef:
            await m.reply_text(reply)
            LOGGER.error(ef)
            LOGGER.error(format_exc())
        return False

    return True


async def owner_check(m) -> bool:
    """Checks if user is owner or not."""
    if isinstance(m, Message):
        user_id = m.from_user.id
    if isinstance(m, CallbackQuery):
        user_id = m.message.from_user.id
        m = m.message

    try:
        if user_id in SUDO_LEVEL:
            return True
    except Exception as ef:
        LOGGER.info(ef, m)
        LOGGER.error(format_exc())

    user = await m.chat.get_member(user_id)

    if user.status != "creator":
        if user.status == "administrator":
            reply = "Stay in your limits, or lose adminship too."
        else:
            reply = "You ain't even admin, what are you trying to do?"
        try:
            await m.edit_text(reply)
        except Exception as ef:
            await m.reply_text(reply)
            LOGGER.error(ef)
            LOGGER.error(format_exc())

        return False

    return True
