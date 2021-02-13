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


from alita import OWNER_ID, DEV_USERS


async def admin_check(c, m) -> bool:
    chat_id = m.chat.id
    user_id = m.from_user.id

    if int(user_id) == int(OWNER_ID) or int(user_id) in DEV_USERS:
        return True

    user = await c.get_chat_member(chat_id=chat_id, user_id=user_id)
    admin_strings = ["creator", "administrator"]

    if user.status not in admin_strings:
        await m.reply_text(
            "This is an Admin Restricted command and you're not allowed to use it.",
        )
        return False

    return True


async def owner_check(c, m) -> bool:
    chat_id = m.chat.id
    user_id = m.from_user.id

    if int(user_id) == int(OWNER_ID) or int(user_id) in DEV_USERS:
        return True

    user = await c.get_chat_member(chat_id=chat_id, user_id=user_id)

    if user.status != "creator":
        await m.reply_text(
            "This is an Owner Restricted command and you're not allowed to use it.",
        )
        return False

    return True
