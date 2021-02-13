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


from html import escape
from pyrogram import filters
from pyrogram.types import Message
from alita import PREFIX_HANDLER, LOGGER
from alita.bot_class import Alita
from alita.db import blacklist_db as db, approve_db as app_db
from alita.utils.localization import GetLang
from alita.utils.regex_utils import regex_searcher
from alita.utils.admin_check import admin_check


__PLUGIN__ = "Blacklist"

__help__ = """
Want to restrict certain words or sentences in your group?

Blacklists are used to stop certain triggers from being said in a group. Any time the trigger is mentioned, \
the message will immediately be deleted. A good combo is sometimes to pair this up with warn filters!
**NOTE:** blacklists do not affect group admins.
 × /blacklist: View the current blacklisted words.
**Admin only:**
 × /addblacklist <triggers>: Add a trigger to the blacklist. Each line is considered one trigger, so using different \
lines will allow you to add muser_listtiple triggers.
 × /unblacklist <triggers>: Remove triggers from the blacklist. Same newline logic applies here, so you can remove \
muser_listtiple triggers at once.
 × /rmblacklist <triggers>: Same as above.

**Note:** Can only remove one remove one blacklist at a time!
"""


@Alita.on_message(filters.command("blacklist", PREFIX_HANDLER) & filters.group)
async def view_blacklist(c: Alita, m: Message):

    res = await admin_check(c, m)
    if not res:
        return

    _ = GetLang(m).strs
    chat_title = m.chat.title
    blacklists_chat = _("blacklist.curr_blacklist_initial").format(
        chat_title=chat_title
    )
    all_blacklisted = db.get_chat_blacklist(m.chat.id)

    if not all_blacklisted:
        await m.reply_text(_("blacklist.no_blacklist").format(chat_title=chat_title))
        return

    for trigger in all_blacklisted:
        blacklists_chat += f" • <code>{escape(trigger)}</code>\n"

    await m.reply_text(blacklists_chat, reply_to_message_id=m.message_id)
    return


@Alita.on_message(filters.command("addblacklist", PREFIX_HANDLER) & filters.group)
async def add_blacklist(c: Alita, m: Message):

    res = await admin_check(c, m)
    if not res:
        return

    _ = GetLang(m).strs
    if len(m.text.split()) >= 2:
        bl_word = m.text.split(None, 1)[1]
        db.add_to_blacklist(m.chat.id, bl_word.lower())
        await m.reply_text(
            _("blacklist.added_blacklist").format(trigger=bl_word),
            reply_to_message_id=m.message_id,
        )
        return
    await m.reply_text(_("general.check_help"), reply_to_message_id=m.message_id)
    return


@Alita.on_message(
    filters.command(["rmblacklist", "unblacklist"], PREFIX_HANDLER) & filters.group
)
async def rm_blacklist(c: Alita, m: Message):

    res = await admin_check(c, m)
    if not res:
        return

    _ = GetLang(m).strs
    chat_bl = db.get_chat_blacklist(m.chat.id)
    if not isinstance(chat_bl, bool):
        pass
    else:
        if len(m.text.split()) >= 2:
            bl_word = m.text.split(None, 1)[1]
            if bl_word in chat_bl:
                db.rm_from_blacklist(m.chat.id, bl_word.lower())
                await m.reply_text(_("blacklist.rm_blacklist").format(bl_word=bl_word))
                return
            await m.reply_text(_("blacklist.no_bl_found").format(bl_word=bl_word))
        else:
            await m.reply_text(
                _("general.check_help"), reply_to_message_id=m.message_id
            )
    return


@Alita.on_message(filters.group, group=11)
async def del_blacklist(c: Alita, m: Message):
    try:
        user_list = []
        approved_users = app_db.all_approved(m.chat.id)
        for auser in approved_users:
            user_list.append(int(auser.user_id))
        async for i in m.chat.iter_members(filter="administrators"):
            user_list.append(i.user.id)
        if m.from_user.id in user_list:
            return
        if m.text:
            chat_filters = db.get_chat_blacklist(m.chat.id)
            if not chat_filters:
                return
            for trigger in chat_filters:
                pattern = r"( |^|[^\w])" + trigger + r"( |$|[^\w])"
                match = await regex_searcher(pattern, m.text.lower())
                if not match:
                    continue
                if match:
                    try:
                        await m.delete()
                    except Exception as ef:
                        LOGGER.info(ef)
                    break
    except AttributeError:
        pass  # Skip attribute errors!
