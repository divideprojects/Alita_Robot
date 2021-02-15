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


from time import time

from pyrogram import errors, filters
from pyrogram.types import Message

from alita import LOGGER, PREFIX_HANDLER, SUPPORT_GROUP
from alita.bot_class import Alita
from alita.utils.admin_check import admin_check
from alita.utils.extract_user import extract_user
from alita.utils.localization import GetLang
from alita.utils.parser import mention_html

__PLUGIN__ = "Bans"
__help__ = """
Someone annoying entered your group?
Want to ban/restriction him/her?
This is the plugin for you, easily kick, ban and unban members on a group.

**Admin only:**
 × /kick: Kixka thw uaer replied or tagged.
 × /ban: Bans the user replied to or tagged.
 × /unban: Unbans the user replied to or tagged.
"""


@Alita.on_message(filters.command("kick", PREFIX_HANDLER) & filters.group)
async def kick_usr(c: Alita, m: Message):

    _ = GetLang(m).strs

    if not (await admin_check(c, m)):
        return

    from_user = await m.chat.get_member(m.from_user.id)

    if from_user.can_restrict_members or from_user.status == "creator":
        user_id, user_first_name = await extract_user(m)
        try:
            await c.kick_chat_member(m.chat.id, user_id, int(time() + 45))
            await m.reply_text(
                f"Banned {(await mention_html(user_first_name, user_id))}",
            )
        except errors.ChatAdminRequired:
            await m.reply_text(_("admin.notadmin"))
        except Exception as ef:
            await m.reply_text(f"<code>{ef}</code>\nReport to @{SUPPORT_GROUP}")
            LOGGER.error(ef)

    return


@Alita.on_message(filters.command("ban", PREFIX_HANDLER) & filters.group)
async def ban_usr(c: Alita, m: Message):

    _ = GetLang(m).strs

    if not (await admin_check(c, m)):
        return

    from_user = await m.chat.get_member(m.from_user.id)

    if from_user.can_restrict_members or from_user.status == "creator":
        user_id, user_first_name = await extract_user(m)
        try:
            await c.kick_chat_member(m.chat.id, user_id)
            await m.reply_text(
                f"Banned {(await mention_html(user_first_name, user_id))}",
            )
        except errors.ChatAdminRequired:
            await m.reply_text(_("admin.notadmin"))
        except Exception as ef:
            await m.reply_text(f"<code>{ef}</code>\nReport to @{SUPPORT_GROUP}")
            LOGGER.error(ef)

    return


@Alita.on_message(filters.command("unban", PREFIX_HANDLER) & filters.group)
async def unban_usr(c: Alita, m: Message):

    _ = GetLang(m).strs

    if not (await admin_check(c, m)):
        return

    from_user = await m.chat.get_member(m.from_user.id)

    if from_user.can_restrict_members or from_user.status == "creator":
        user_id, user_first_name = await extract_user(m)
        try:
            await c.unban_chat_member(m.chat.id, user_id)
            await m.reply_text(
                f"Unbanned {(await mention_html(user_first_name, user_id))}",
            )
        except errors.ChatAdminRequired:
            await m.reply_text(_("admin.notadmin"))
        except Exception as ef:
            await m.reply_text(f"<code>{ef}</code>\nReport to @{SUPPORT_GROUP}")
            LOGGER.error(ef)

    return
