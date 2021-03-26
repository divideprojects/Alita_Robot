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
from pyrogram.types import CallbackQuery, Message

from alita import DEV_USERS, OWNER_ID, SUDO_USERS
from alita.bot_class import Alita
from alita.tr_engine import tlang
from alita.utils.caching import ADMIN_CACHE, admin_cache_reload

SUDO_LEVEL = set(SUDO_USERS + DEV_USERS + [int(OWNER_ID)])
DEV_LEVEL = set(DEV_USERS + [int(OWNER_ID)])


import re
import shlex
from typing import List


def command(
    commands: str or List[str],
    prefixes: str or List[str] = "/",
    case_sensitive: bool = False,
):
    from alita import BOT_USERNAME

    async def func(flt, _: Alita, message: Message):
        text: str = message.text or message.caption
        message.command = None
        if not text:
            return False
        regex = "^({prefix})+\\b({regex})\\b(\\b@{bot_name}\\b)?(.*)".format(
            prefix="|".join(re.escape(x) for x in flt.prefixes),
            regex="|".join(flt.commands),
            bot_name=BOT_USERNAME,
        )
        matches = re.search(re.compile(regex), text)
        if matches:
            message.command = [matches.group(2)]
            for arg in shlex.split(matches.group(4).strip()):
                message.command.append(arg)
            return True
        else:
            return False

    commands = commands if type(commands) is list else [commands]
    commands = {c if case_sensitive else c.lower() for c in commands}
    prefixes = [] if prefixes is None else prefixes
    prefixes = prefixes if type(prefixes) is list else [prefixes]
    prefixes = set(prefixes) if prefixes else {""}
    return filters.create(
        func,
        "CustomCommandFilter",
        commands=commands,
        prefixes=prefixes,
        case_sensitive=case_sensitive,
    )


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

    if m.chat.type != "supergroup":
        return False

    # Bypass the bot devs, sudos and owner
    if m.from_user.id in SUDO_LEVEL:
        return True

    try:
        admin_group = {i[0] for i in ADMIN_CACHE[m.chat.id]}
    except KeyError:
        admin_group = {
            i[0] for i in await admin_cache_reload(m, "custom_filter_update")
        }
    except ValueError as ef:
        # To make language selection work in private chat of user, i.e. PM
        if ("The chat_id" and "belongs to a user") in ef:
            return True

    if m.from_user.id in admin_group:
        return True

    await m.reply_text(tlang(m, "general.no_admin_cmd_perm"))

    return False


async def owner_check_func(_, __, m):
    """Check if user is Owner or not."""

    if isinstance(m, CallbackQuery):
        m = m.message

    if m.chat.type != "supergroup":
        return False

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
            msg = "Do you think that you can execute owner commands?"
        await m.reply_text(msg)

    return status


async def restrict_check_func(_, __, m):
    """Check if user can restrict users or not."""

    if isinstance(m, CallbackQuery):
        m = m.message

    if m.chat.type != "supergroup":
        return False

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

    if m.chat.type != "supergroup":
        return False

    # Bypass the bot devs, sudos and owner
    if m.from_user.id in DEV_LEVEL:
        return True

    user = await m.chat.get_member(m.from_user.id)

    if user.can_promote_members or user.status == "creator":
        status = True
    else:
        status = False
        await m.reply_text(tlang(m, "admin.promote.no_promote_perm"))

    return status


async def invite_check_func(_, __, m):
    """Check if user can invite users or not."""

    if isinstance(m, CallbackQuery):
        m = m.message

    if m.chat.type != "supergroup":
        return False

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
