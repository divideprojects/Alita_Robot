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
from pyrogram.errors import ChatAdminRequired, RightForbidden, RPCError
from pyrogram.types import ChatPermissions, Message

from alita import LOGGER, PREFIX_HANDLER, SUPPORT_GROUP, SUPPORT_STAFF
from alita.bot_class import Alita
from alita.tr_engine import tlang
from alita.utils.custom_filters import restrict_filter
from alita.utils.extract_user import extract_user
from alita.utils.parser import mention_html

__PLUGIN__ = "Muting"
__help__ = """
Want someone to keep quite for a while in the group?
Mute plugin is here to help, mute or unmute any user easily!

**Admin only:**
 × /mute: Mute the user replied to or mentioned.
 × /unmute: Unmutes the user mentioned or replied to.

An example of promoting someone to admins:
`/mute @username`; this mutes a user.
"""


@Alita.on_message(
    filters.command("mute", PREFIX_HANDLER) & filters.group & restrict_filter,
)
async def mute_usr(c: Alita, m: Message):

    user_id, user_first_name = await extract_user(c, m)

    if user_id in SUPPORT_STAFF:
        await m.reply_text("This user is in my support staff, cannot restrict them.")
        return

    try:
        await m.chat.restrict_member(
            user_id,
            ChatPermissions(
                can_send_messages=False,
                can_send_media_messages=False,
                can_send_stickers=False,
                can_send_animations=False,
                can_send_games=False,
                can_use_inline_bots=False,
                can_add_web_page_previews=False,
                can_send_polls=False,
                can_change_info=False,
                can_invite_users=True,
                can_pin_messages=False,
            ),
        )
        await m.reply_text(
            (await tlang(m, "admin.muted_user")).format(
                user=(await mention_html(user_first_name, user_id)),
            ),
        )
    except ChatAdminRequired:
        await m.reply_text(await tlang(m, "admin.not_admin"))
    except RightForbidden:
        await m.reply_text(await tlang(m, "admin.bot_no_mute_right"))
    except RPCError as ef:
        await m.reply_text(
            (await tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=f"@{SUPPORT_GROUP}",
                ef=f"<code>{ef}</code>",
            ),
        )
        LOGGER.error(ef)

    return


@Alita.on_message(
    filters.command("unmute", PREFIX_HANDLER) & filters.group & restrict_filter,
)
async def unmute_usr(c: Alita, m: Message):

    user_id, user_first_name = await extract_user(c, m)

    try:
        await m.chat.restrict_member(user_id, m.chat.permissions)
        await m.reply_text(
            (await tlang(m, "admin.unmuted_user")).format(
                user=(await mention_html(user_first_name, user_id)),
            ),
        )
    except ChatAdminRequired:
        await m.reply_text(await tlang(m, "admin.not_admin"))
    except RightForbidden:
        await m.reply_text(await tlang(m, "admin.bot_no_mute_right"))
    except RPCError as ef:
        await m.reply_text(
            (await tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=f"@{SUPPORT_GROUP}",
                ef=f"<code>{ef}</code>",
            ),
        )
        LOGGER.error(ef)
    return
