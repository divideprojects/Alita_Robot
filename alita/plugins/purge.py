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


from asyncio import sleep

from pyrogram import filters
from pyrogram.errors import MessageDeleteForbidden, RPCError
from pyrogram.types import Message

from alita import PREFIX_HANDLER, SUPPORT_GROUP
from alita.bot_class import Alita
from alita.tr_engine import tlang
from alita.utils.custom_filters import admin_filter


@Alita.on_message(filters.command("purge", PREFIX_HANDLER) & admin_filter)
async def purge(c: Alita, m: Message):

    if m.chat.type != "supergroup":
        await m.reply_text(tlang(m, "purge.err_basic"))
        return

    if m.reply_to_message:
        message_ids = list(range(m.reply_to_message.message_id, m.message_id))

        def divide_chunks(l, n):
            for i in range(0, len(l), n):
                yield l[i : i + n]

        # Dielete messages in chunks of 100 messages
        m_list = list(divide_chunks(message_ids, 100))

        try:
            for plist in m_list:
                await c.delete_messages(
                    chat_id=m.chat.id,
                    message_ids=plist,
                    revoke=True,
                )
            await m.delete()
        except MessageDeleteForbidden:
            await m.reply_text(tlang(m, "purge.old_msg_err"))
            return
        except RPCError as ef:
            await m.reply_text(
                (tlang(m, "general.some_error")).format(
                    SUPPORT_GROUP=SUPPORT_GROUP,
                    ef=ef,
                ),
            )

        count_del_msg = len(message_ids)

        z = await m.reply_text(
            (tlang(m, "purge.purge_msg_count")).format(
                msg_count=count_del_msg,
            ),
        )
        await sleep(3)
        await z.delete()
        return
    await m.reply_text("Reply to a message to start purge.")
    return


@Alita.on_message(
    filters.command("del", PREFIX_HANDLER) & admin_filter,
    group=9,
)
async def del_msg(c: Alita, m: Message):

    if m.chat.type != "supergroup":
        return

    if m.reply_to_message:
        await m.delete()
        await c.delete_messages(
            chat_id=m.chat.id,
            message_ids=m.reply_to_message.message_id,
        )
    else:
        await m.reply_text(tlang(m, "purge.what_del"))
    return


__PLUGIN__ = "plugins.purges.main"
__help__ = "plugins.purges.help"
__alt_name__ = ["purge", "del"]
