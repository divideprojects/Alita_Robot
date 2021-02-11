import asyncio
from alita.__main__ import Alita
from pyrogram import filters, errors
from pyrogram.types import Message
from alita import PREFIX_HANDLER
from alita.utils.localization import GetLang
from alita.utils.admin_check import admin_check


__PLUGIN__ = "Purges"

__help__ = """
Want to delete messages in you group?

 -/purge: Deletes messages upto replied message.
 -/purge <X>: Delete the number of messages specifed by number X
 × /del: Deletes a single message, used as a reply to message.
"""


@Alita.on_message(filters.command("purge", PREFIX_HANDLER) & filters.group)
async def purge(c: Alita, m: Message):

    res = await admin_check(c, m)
    if not res:
        return

    _ = await GetLang(m).strs
    if m.chat.type != "supergroup":
        await m.reply_text(_("purge.err_basic"))
        return
    dm = await m.reply_text(_("purge.deleting"))

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
    except errors.MessageDeleteForbidden:
        await dm.edit_text(_("purge.old_msg_err"))
        return

    count_del_msg = len(message_ids)

    await dm.edit(_("purge.purge_msg_count").format(msg_count=count_del_msg))
    await asyncio.sleep(3)
    await dm.delete()
    return


@Alita.on_message(filters.command("del", PREFIX_HANDLER) & filters.group, group=3)
async def del_msg(c: Alita, m: Message):

    res = await admin_check(c, m)
    if not res:
        return

    _ = await GetLang(m).strs
    if m.reply_to_message:
        if m.chat.type != "supergroup":
            return
        await c.delete_messages(
            chat_id=m.chat.id, message_ids=m.reply_to_message.message_id
        )
        await asyncio.sleep(0.5)
        await m.delete()
    else:
        await m.reply_text(_("purge.what_del"))
    return
