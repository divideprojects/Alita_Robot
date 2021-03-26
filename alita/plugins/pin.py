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
from alita.database.antichannelpin_db import Pins
from alita.tr_engine import tlang
from alita.utils.custom_filters import admin_filter

# Initialize
pinsdb = Pins()


@Alita.on_message(filters.command("pin", PREFIX_HANDLER) & admin_filter)
async def pin_message(_, m: Message):

    pin_args = m.text.split(None, 1)
    if m.reply_to_message:
        try:
            disable_notification = True

            if len(pin_args) >= 2 and pin_args[1] in ["alert", "notify", "loud"]:
                disable_notification = False

            await m.reply_to_message.pin(
                disable_notification=disable_notification,
            )
            LOGGER.info(
                f"{m.from_user.id} pinned msgid-{m.reply_to_message.message_id} in {m.chat.id}",
            )
            if (str(m.chat.id)).startswith("-100"):
                link_chat_id = (str(m.chat.id)).replace("-100", "")
            message_link = (
                f"https://t.me/c/{link_chat_id}/{m.reply_to_message.message_id}"
            )
            await m.reply_text(
                tlang(m, "pin.pinned_msg").format(message_link=message_link),
            )

        except ChatAdminRequired:
            await m.reply_text(tlang(m, "admin.not_admin"))
        except RightForbidden:
            await m.reply_text(tlang(m, "pin.no_rights_pin"))
        except RPCError as ef:
            await m.reply_text(
                (tlang(m, "general.some_error")).format(
                    SUPPORT_GROUP=SUPPORT_GROUP,
                    ef=ef,
                ),
            )
            LOGGER.error(ef)
    else:
        await m.reply_text("Reply to a message to pin it!")

    return


@Alita.on_message(filters.command("unpin", PREFIX_HANDLER) & admin_filter)
async def unpin_message(c: Alita, m: Message):

    try:
        if m.reply_to_message:
            await c.unpin_chat_message(m.chat.id, m.reply_to_message.message_id)
            LOGGER.info(
                f"{m.from_user.id} unpinned msgid-{m.reply_to_message.message_id} in {m.chat.id}",
            )
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
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)

    return


@Alita.on_message(filters.command("unpinall", PREFIX_HANDLER) & admin_filter)
async def unpinall_message(c: Alita, m: Message):

    try:
        await c.unpin_all_chat_messages(m.chat.id)
        LOGGER.info(f"{m.from_user.id} unpinned all messages in {m.chat.id}")
        await m.reply_text(tlang(m, "pin.unpinned_all_msg"))
    except ChatAdminRequired:
        await m.reply_text(tlang(m, "admin.notadmin"))
    except RightForbidden:
        await m.reply_text(tlang(m, "pin.no_rights_unpin"))
    except RPCError as ef:
        await m.reply_text(
            (tlang(m, "general.some_error")).format(
                SUPPORT_GROUP=SUPPORT_GROUP,
                ef=ef,
            ),
        )
        LOGGER.error(ef)

    return


@Alita.on_message(filters.command("antichannelpin", PREFIX_HANDLER) & admin_filter)
async def anti_channel_pin(_, m: Message):

    if len(m.text.split()) == 1:
        status = pinsdb.check_status(m.chat.id, "antichannelpin")
        await m.reply_text(
            tlang(m, "pin.antichannelpin.current_status").format(
                status=status,
            ),
        )
        return

    if len(m.text.split()) == 2:
        if m.command[1] in ("yes", "on", "false"):
            pinsdb.set_on(m.chat.id, "antichannelpin")
            LOGGER.info(f"{m.from_user.id} enabled antichannelpin in {m.chat.id}")
            msg = tlang(m, "pin.antichannelpin.turned_on")
        elif m.command[1] in ("no", "off", "true"):
            pinsdb.set_on(m.chat.id, "antichannelpin")
            LOGGER.info(f"{m.from_user.id} disabled antichannelpin in {m.chat.id}")
            msg = tlang(m, "pin.antichannelpin.turned_off")
        else:
            await m.reply_text(tlang(m, "pin.general.check_help"))
            return

    await m.reply_text(msg)
    return


@Alita.on_message(filters.command("cleanlinked", PREFIX_HANDLER) & admin_filter)
async def clean_linked(_, m: Message):

    if len(m.text.split()) == 1:
        status = pinsdb.check_status(m.chat.id, "cleanlinked")
        await m.reply_text(
            tlang(m, "pin.antichannelpin.current_status").format(
                status=status,
            ),
        )
        return

    if len(m.text.split()) == 2:
        if m.command[1] in ("yes", "on", "false"):
            pinsdb.set_on(m.chat.id, "cleanlinked")
            LOGGER.info(f"{m.from_user.id} enabled CleanLinked in {m.chat.id}")
            msg = "Turned on CleanLinked! Now all the messages from linked channel will be deleted!"
        elif m.command[1] in ("no", "off", "true"):
            pinsdb.set_on(m.chat.id, "cleanlinked")
            LOGGER.info(f"{m.from_user.id} disabled CleanLinked in {m.chat.id}")
            msg = "Turned off CleanLinked! Messages from linked channel will not be deleted!"
        else:
            await m.reply_text(tlang(m, "pin.general.check_help"))
            return

    await m.reply_text(msg)
    return


@Alita.on_message(filters.command("permapin", PREFIX_HANDLER) & admin_filter)
async def perma_pin(_, m: Message):
    if m.reply_to_message:
        LOGGER.info(f"{m.from_user.id} used permampin in {m.chat.id}")
        z = await m.reply_to_message.copy(m.chat.id)
        await z.pin()
    elif len(m.text.split()) > 1:
        z = await m.reply_text(m.text.split(None, 1)[1])
        await z.pin()
    else:
        await m.reply_text("Reply to a message or enter text to pin it.")

    return


__PLUGIN__ = "plugins.pins.main"
__help__ = "plugins.pins.help"
__alt_name__ = ["pin", "unpin"]
