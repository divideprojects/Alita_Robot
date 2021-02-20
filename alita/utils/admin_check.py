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


from alita import DEV_USERS, OWNER_ID


async def admin_check(m) -> bool:
    """Checks if user is admin or not."""
    user_id = m.from_user.id

    if (int(user_id) == int(OWNER_ID)) or (int(user_id) in DEV_USERS):
        return True

    user = await m.chat.get_member(user_id)
    admin_strings = ("creator", "administrator")

    if user.status not in admin_strings:
        await m.reply_text(
            "Nigga, you're not admin, don't try this explosive shit.",
        )
        return False

    return True


async def owner_check(m) -> bool:
    """Checks if user is owner or not."""
    user_id = m.from_user.id

    if (int(user_id) == int(OWNER_ID)) or (int(user_id) in DEV_USERS):
        return True

    user = await m.chat.get_member(user_id)

    if user.status != "creator":
        if user.status == "administrator":
            reply = "Stay in your limits, or lose adminship too."
        else:
            reply = "You ain't even admin, what are you trying to do?"
        await m.reply_text(reply)
        return False

    return True
