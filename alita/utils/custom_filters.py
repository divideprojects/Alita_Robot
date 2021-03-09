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


from pyrogram import filters
from pyrogram.types import CallbackQuery

from alita import DEV_USERS, OWNER_ID, SUDO_USERS
from alita.tr_engine import tlang

from time import perf_counter
from cachetools import TTLCache

from threading import RLock

ADMIN_CACHE = TTLCache(maxsize=512, ttl=60 * 10, timer=perf_counter)
THREAD_LOCK = RLock()

SUDO_LEVEL = SUDO_USERS + DEV_USERS + [int(OWNER_ID)]
DEV_LEVEL = DEV_USERS + [int(OWNER_ID)]


async def dev_check_func(_, __, m):
    """Check if user is Dev or not."""
    return bool(m.from_user.id in DEV_USERS or m.from_user.id == int(OWNER_ID))


async def sudo_check_func(_, __, m):
    """Check if user is Sudo or not."""
    return bool(m.from_user.id in SUDO_LEVEL)


async def admin_check_func(_, __, m):
    """Check if user is Admin or not."""
    if isinstance(m, CallbackQuery):
        m = m.message

    # Bypass the bot devs, sudos and owner
    if m.from_user.id in SUDO_LEVEL:
        return True

    with THREAD_LOCK:
        try:
            admin_list = [user[0] for user in ADMIN_CACHE[m.chat.id]]
            if m.from_user.id in admin_list:
                return True
            else:
                await m.reply_text(tlang(m, "general.no_admin_cmd_perm"))
                return False
        except KeyError;
            admins_list = []
            async for i in m.chat.iter_members(filter="administrators"):
                admins_list.append((i.user.id, ("@"+i.user.username) if i.user.username else i.user.first_name))
            ADMIN_CACHE[m.chat.id] = admins_list
            if m.from_user.id in admin_list:
                return True
            else:
                await m.reply_text(tlang(m, "general.no_admin_cmd_perm"))
                return False
        except ValueError as ef:  # To make language selection work in private chat of user, i.e. PM
            if ("The chat_id" or "belongs to a user") in ef:
                return True
        except Exception as ef:
            user = await m.chat.get_member(m.from_user.id)
            if user.status in ("creator", "administrator"):
                return True

    return False


async def owner_check_func(_, __, m):
    """Check if user is Owner or not."""
    if isinstance(m, CallbackQuery):
        m = m.message

    # Bypass the bot devs, sudos and owner
    if m.from_user.id in DEV_LEVEL:
        return True
    user = await m.chat.get_member(m.from_user.id)

    if user.status == "creator":
        status = True
    else:
        status = False
        if user.status == "administrator":
            msg = "You're an admin only, stay in your limits!"
        else:
            msg = "Do you think that you can execute admin commands?"
        await m.reply_text(msg)

    return status


async def restrict_check_func(_, __, m):
    """Check if user can restrict users or not."""
    if isinstance(m, CallbackQuery):
        m = m.message

    # Bypass the bot devs, sudos and owner
    if m.from_user.id in DEV_LEVEL:
        return True
    user = await m.chat.get_member(m.from_user.id)

    if user.can_restrict_members or user.status == "creator":
        status = True
    else:
        status = False
        await m.reply_text(tlang(m, "admin.no_restrict_perm"))

    return status


async def promote_check_func(_, __, m):
    """Check if user can promote users or not."""
    if isinstance(m, CallbackQuery):
        m = m.message

    # Bypass the bot devs, sudos and owner
    if m.from_user.id in DEV_LEVEL:
        return True
    user = await m.chat.get_member(m.from_user.id)

    if user.can_promote_members or user.status == "creator":
        status = True
    else:
        status = False
        await m.reply_text(tlang(m, "admin.no_promote_demote_perm"))

    return status


async def invite_check_func(_, __, m):
    """Check if user can invite users or not."""
    if isinstance(m, CallbackQuery):
        m = m.message

    # Bypass the bot devs, sudos and owner
    if m.from_user.id in DEV_LEVEL:
        return True

    user = await m.chat.get_member(m.from_user.id)

    if user.can_invite_users or user.status == "creator":
        status = True
    else:
        status = False
        await m.reply_text(tlang(m, "admin.no_user_invite_perm"))

    return status


dev_filter = filters.create(dev_check_func)
sudo_filter = filters.create(sudo_check_func)
admin_filter = filters.create(admin_check_func)
owner_filter = filters.create(owner_check_func)
restrict_filter = filters.create(restrict_check_func)
promote_filter = filters.create(promote_check_func)
invite_filter = filters.create(invite_check_func)
