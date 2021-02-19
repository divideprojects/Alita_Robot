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
from alita.utils.custom_filters import admin_filter
from alita.utils.localization import GetLang

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

    _ = GetLang(m).strs

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
            await m.reply_text(_("admin.pinnedmsg"))

        except ChatAdminRequired:
            await m.reply_text(_("admin.notadmin"))
        except RightForbidden:
            await m.reply_text("I don't have enough rights to pin messages.")
        except RPCError as ef:
            await m.reply_text(f"<code>{ef}</code>\nReport to @{SUPPORT_GROUP}")
            LOGGER.error(ef)
    else:
        await m.reply_text(_("admin.nopinmsg"))

    return


@Alita.on_message(filters.command("unpin", PREFIX_HANDLER) & filters.group & admin_filter)
async def unpin_message(c: Alita, m: Message):

    _ = GetLang(m).strs

    try:
        await c.unpin_chat_message(m.chat.id)
        await m.reply_text("Unpinned last message.")
    except ChatAdminRequired:
        await m.reply_text(_("admin.notadmin"))
    except RightForbidden:
        await m.reply_text("I don't have enough rights to unpin messages")
    except RPCError as ef:
        await m.reply_text(f"<code>{ef}</code>\nReport to @{SUPPORT_GROUP}")
        LOGGER.error(ef)

    return


@Alita.on_message(filters.command("unpinall", PREFIX_HANDLER) & filters.group & admin_filter)
async def unpinall_message(c: Alita, m: Message):

    _ = GetLang(m).strs

    try:
        await c.unpin_all_chat_messages(m.chat.id)
        await m.reply_text("Unpinned all messages in this chat.")
    except ChatAdminRequired:
        await m.reply_text(_("admin.notadmin"))
    except RightForbidden:
        await m.reply_text("I don't have enough rights to unpin messages")
    except RPCError as ef:
        await m.reply_text(f"<code>{ef}</code>\nReport to @{SUPPORT_GROUP}")
        LOGGER.error(ef)

    return
