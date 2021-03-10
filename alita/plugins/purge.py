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
from pyrogram.errors import MessageDeleteForbidden
from pyrogram.types import Message

from alita import PREFIX_HANDLER
from alita.bot_class import Alita
from alita.tr_engine import tlang
from alita.utils.custom_filters import admin_filter

__PLUGIN__ = "plugins.purges.main"
__help__ = "plugins.purges.help"


@Alita.on_message(
    filters.command("purge", PREFIX_HANDLER) & filters.group & admin_filter,
)
async def purge(c: Alita, m: Message):

    if m.chat.type != "supergroup":
        await m.reply_text(tlang(m, "purge.err_basic"))
        return
    dm = await m.reply_text(tlang(m, "purge.deleting"))

    message_ids = []

    if m.reply_to_message:
        for a_msg in range(m.reply_to_message.message_id, m.message_id):
            message_ids.append(a_msg)

    if (
        not m.reply_to_message
        and len(m.text.split()) == 2
        and isinstance(m.text.split()[1], int)
    ):
        c_msg_id = m.message_id
        first_msg = (m.message_id) - (m.text.split()[1])
        for a_msg in range(first_msg, c_msg_id):
            message_ids.append(a_msg)

    try:
        await c.delete_messages(chat_id=m.chat.id, message_ids=message_ids, revoke=True)
        await m.delete()
    except MessageDeleteForbidden:
        await dm.edit_text(tlang(m, "purge.old_msg_err"))
        return

    count_del_msg = len(message_ids)

    await dm.edit(
        (tlang(m, "purge.purge_msg_count")).format(
            msg_count=count_del_msg,
        ),
    )
    await sleep(3)
    await dm.delete()
    return


@Alita.on_message(
    filters.command("del", PREFIX_HANDLER) & filters.group & admin_filter,
    group=4,
)
async def del_msg(c: Alita, m: Message):

    if m.reply_to_message:
        if m.chat.type != "supergroup":
            return
        await c.delete_messages(
            chat_id=m.chat.id,
            message_ids=m.reply_to_message.message_id,
        )
        await m.delete()
    else:
        await m.reply_text(tlang(m, "purge.what_del"))
    return
