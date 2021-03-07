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
from pyrogram.types import Message

from alita import LOGGER, PREFIX_HANDLER, SUPPORT_GROUP
from alita.bot_class import Alita
from alita.tr_engine import tlang
from alita.utils.custom_filters import admin_filter

__PLUGIN__ = "Pins"
__help__ = """
Here you find find all help related to groups pins and how to manage them via me.

**Admin Cmds:**
 × /pin: Silently pins the message replied to - add `loud`, `notify` or `alert` to give notificaton to users.
 × /unpin: Unpins the last pinned message.
 × /unpinall: Unpins all the pinned message in the current chat.
"""


@Alita.on_message(filters.command("pin", PREFIX_HANDLER) & filters.group & admin_filter)
async def pin_message(c: Alita, m: Message):

    pin_args = m.text.split(None, 1)
    if m.reply_to_message:
        try:
            disable_notification = True

            if len(pin_args) >= 2 and pin_args[1] in ["alert", "notify", "loud"]:
                disable_notification = False

            await c.pin_chat_message(
                m.chat.id,
                m.reply_to_message.message_id,
                disable_notification=disable_notification,
            )
            await m.reply_text(tlang(m, "pin.pinned_msg"))

        except ChatAdminRequired:
            await m.reply_text(tlang(m, "admin.not_admin"))
        except RightForbidden:
            await m.reply_text(tlang(m, "pin.no_rights_pin"))
        except RPCError as ef:
            await m.reply_text(
                (tlang(m, "general.some_error")).format(
                    SUPPORT_GROUP=f"@{SUPPORT_GROUP}",
                    ef=f"<code>{ef}</code>",
                ),
            )
            LOGGER.error(ef)
    else:
        await m.reply_text(tlang(m, "admin.nopinmsg"))

    return


@Alita.on_message(
    filters.command("unpin", PREFIX_HANDLER) & filters.group & admin_filter,
)
async def unpin_message(c: Alita, m: Message):

    try:
        if m.reply_to_message:
            await c.unpin_chat_message(m.chat.id, m.reply_to_message.message_id)
            await m.reply_text(tlang(m, "pin.unpinned_last_msg"))
        else:
            await m.reply_text(tlang(m, "pin.reply_to_unpin"))
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.not_admin"))
    except RightForbidden:
        await m.reply_text(tlang(m, "pin.no_rights_unpin"))
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=f"@{SUPPORT_GROUP}",
                ef=f"<code>{ef}</code>",
            ),
        )
        LOGGER.error(ef)

    return


@Alita.on_message(
    filters.command("unpinall", PREFIX_HANDLER) & filters.group & admin_filter,
)
async def unpinall_message(c: Alita, m: Message):

    try:
        await c.unpin_all_chat_messages(m.chat.id)
        await m.reply_text(tlang(m, "pin.unpinned_all_msg"))
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.notadmin"))
    except RightForbidden:
        await m.reply_text(tlang(m, "pin.no_rights_unpin"))
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=f"@{SUPPORT_GROUP}",
                ef=f"<code>{ef}</code>",
            ),
        )
        LOGGER.error(ef)

    return
