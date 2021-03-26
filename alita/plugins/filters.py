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


from re import escape as re_escape
from secrets import choice
from traceback import format_exc

from pyrogram import filters
from pyrogram.errors import RPCError
from pyrogram.types import (
    CallbackQuery,
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    Message,
)

from alita import PREFIX_HANDLER
from alita.bot_class import LOGGER, Alita
from alita.database.filters_db import Filters
from alita.utils.cmd_senders import send_cmd
from alita.utils.custom_filters import admin_filter, owner_filter
from alita.utils.msg_types import Types, get_filter_type
from alita.utils.parser import mention_html
from alita.utils.regex_utils import regex_searcher
from alita.utils.string import (
    escape_mentions_using_curly_brackets,
    parse_button,
    split_quotes,
)

# Initialise
db = Filters()


@Alita.on_message(filters.command("filters", PREFIX_HANDLER) & filters.group)
async def view_filters(_, m: Message):

    LOGGER.info(f"{m.from_user.id} checking filters in {m.chat.id}")

    filters_chat = f"Filters in <b>{m.chat.title}</b>:\n"
    all_filters = db.get_all_filters(m.chat.id)
    actual_filters = [j for i in all_filters for j in i.split("|")]

    if not actual_filters:
        await m.reply_text(f"There are no filters in {m.chat.title}")
        return

    filters_chat += "\n".join(
        [
            f" • {' | '.join([f'<code>{i}</code>' for i in i.split('|')])}"
            for i in all_filters
        ],
    )
    await m.reply_text(filters_chat)
    return


@Alita.on_message(
    filters.command(["addfilter", "filter"], PREFIX_HANDLER) & admin_filter,
)
async def add_filter(_, m: Message):

    args = m.text.split(None, 1)
    all_filters = db.get_all_filters(m.chat.id)
    actual_filters = {j for i in all_filters for j in i.split("|")}

    if (len(all_filters) >= 50) and (len(actual_filters) >= 120):
        await m.reply_text(
            "Only 50 filters and 120 aliases are allowed per chat!\nTo  add more filters, remove the existing ones.",
        )
        return

    if not m.reply_to_message and len(m.text.split()) < 3:
        await m.reply_text(
            "Please provide keyboard keyword for this filter to reply with!",
        )
        return

    if m.reply_to_message and len(args) < 2:
        await m.reply_text("Please provide keyword for this filter to reply with!")
        return

    extracted = await split_quotes(args[1])
    keyword = extracted[0].lower()

    for k in keyword.split("|"):
        if k in actual_filters:
            await m.reply_text(f"Filter <code>{k}</code> already exists!")
            return

    teks, msgtype, file_id = await get_filter_type(m)

    if not m.reply_to_message and len(m.text.split()) >= 2:
        teks, _ = await parse_button(extracted[1])
        if not teks:
            await m.reply_text(
                "There is no filter message - You can't JUST have buttons, you need a message to go with it!",
            )
            return

    elif m.reply_to_message and len(m.text.split()) >= 2:
        if m.reply_to_message.text:
            text_to_parsing = m.reply_to_message.text
        elif m.reply_to_message.caption:
            text_to_parsing = m.reply_to_message.caption
        else:
            text_to_parsing = ""
        teks, _ = await parse_button(text_to_parsing)

    elif not teks and not msgtype:
        await m.reply_text(
            'Please provide keyword for this filter reply with!\nEnclose filter in <code>"double quotes"</code>',
        )
        return

    elif m.reply_to_message:

        if m.reply_to_message.text:
            text_to_parsing = m.reply_to_message.text
        elif m.reply_to_message.caption:
            text_to_parsing = m.reply_to_message.caption
        else:
            text_to_parsing = ""

        teks, _ = await parse_button(text_to_parsing)
        if (m.reply_to_message.text or m.reply_to_message.caption) and not teks:
            await m.reply_text(
                "There is no filter message - You can't JUST have buttons, you need a message to go with it!",
            )
            return

    else:
        await m.reply_text("Invalid filter!")
        return

    add = db.save_filter(m.chat.id, keyword, teks, msgtype, file_id)
    LOGGER.info(f"{m.from_user.id} added new filter ({keyword}) in {m.chat.id}")
    if add:
        await m.reply_text(
            f"Saved filter for '<code>{', '.join(keyword.split('|'))}</code>' in <b>{m.chat.title}</b>!",
        )
    await m.stop_propagation()


@Alita.on_message(
    filters.command(["rmfilter", "stop", "unfilter"], PREFIX_HANDLER) & admin_filter,
)
async def stop_filter(_, m: Message):
    args = m.command

    if len(args) < 2:
        await m.reply_text("What should I stop replying to?")
        return

    chat_filters = db.get_all_filters(m.chat.id)
    act_filters = {j for i in chat_filters for j in i.split("|")}

    if not chat_filters:
        await m.reply_text("No filters active here!")
        return

    for keyword in act_filters:
        if keyword == args[1]:
            db.rm_filter(m.chat.id, args[1])
            LOGGER.info(f"{m.from_user.id} removed filter ({keyword}) in {m.chat.id}")
            await m.reply_text(
                f"Okay, I'll stop replying to that filter and it's aliases in <b>{m.chat.title}</b>.",
            )
            await m.stop_propagation()

    await m.reply_text(
        "That's not a filter - Click: /filters to get currently active filters.",
    )
    await m.stop_propagation()


@Alita.on_message(
    filters.command(
        ["rmallfilters", "removeallfilters", "stopall", "stopallfilters"],
        PREFIX_HANDLER,
    )
    & owner_filter,
)
async def rm_allfilters(_, m: Message):

    all_bls = db.get_all_filters(m.chat.id)
    if not all_bls:
        await m.reply_text("No filters to stop in this chat.")
        return

    await m.reply_text(
        "Are you sure you want to clear all filters?",
        reply_markup=InlineKeyboardMarkup(
            [
                [
                    InlineKeyboardButton(
                        "⚠️ Confirm",
                        callback_data=f"rm_allfilters.{m.from_user.id}.{m.from_user.first_name}",
                    ),
                    InlineKeyboardButton("❌ Cancel", callback_data="close"),
                ],
            ],
        ),
    )
    return


@Alita.on_callback_query(filters.regex("^rm_allfilters."))
async def rm_allfilters_callback(_, q: CallbackQuery):
    user_id = q.data.split(".")[-2]
    name = q.data.split(".")[-1]
    user_status = (await q.message.chat.get_member(user_id)).status
    if user_status != "creator":
        await q.message.edit(
            (
                f"You're an admin {await mention_html(name, user_id)}, not owner!\n"
                "Stay in your limits!"
            ),
        )
        return
    db.rm_all_filters(q.message.chat.id)
    await q.message.edit_text("Cleared all filters for {q.message.chat.id}")
    LOGGER.info(f"{user_id} removed all filter from {q.message.chat.id}")
    await q.answer("Cleared all Filters!", show_alert=True)
    return


async def send_filter_reply(c: Alita, m: Message, trigger: str):
    """Reply with assigned filter for the trigger"""
    getfilter = db.get_filter(m.chat.id, trigger)

    if not getfilter:
        await m.reply_text(
            "<b>Error:</b> Cannot find a type for this filter!!",
            quote=True,
        )
        return

    msgtype = getfilter["msgtype"]
    if not msgtype:
        await m.reply_text("<b>Error:</b> Cannot find a type for this filter!!")
        return

    try:
        # support for random filter texts
        splitter = "%%%"
        filter_reply = getfilter["filter_reply"].split(splitter)
        filter_reply = choice(filter_reply)
    except KeyError:
        filter_reply = ""

    parse_words = [
        "first",
        "last",
        "fullname",
        "username",
        "id",
        "chatname",
        "mention",
    ]
    text = await escape_mentions_using_curly_brackets(m, filter_reply, parse_words)

    if msgtype == Types.TEXT:
        await m.reply_text(text, quote=True)

    elif msgtype in (
        Types.STICKER,
        Types.VIDEO_NOTE,
        Types.CONTACT,
        Types.ANIMATED_STICKER,
    ):
        await (await send_cmd(c, msgtype))(
            m.chat.id,
            getfilter["fileid"],
            reply_to_message_id=m.message_id,
        )
    else:
        await (await send_cmd(c, msgtype))(
            m.chat.id,
            getfilter["fileid"],
            caption=text,
            reply_to_message_id=m.message_id,
        )
    return msgtype


@Alita.on_message(filters.text & filters.group, group=6)
async def filters_watcher(c: Alita, m: Message):

    if not m.from_user:
        return

    chat_filters = db.get_all_filters(m.chat.id)
    actual_filters = {j for i in chat_filters for j in i.split("|")}

    for trigger in actual_filters:
        pattern = r"( |^|[^\w])" + re_escape(trigger) + r"( |$|[^\w])"
        match = await regex_searcher(pattern, m.text.lower())
        if match:
            try:
                msgtype = await send_filter_reply(c, m, trigger)
                LOGGER.info(f"Replied with {msgtype} to {trigger} in {m.chat.id}")
            except RPCError as ef:
                LOGGER.error(ef)
                LOGGER.error(format_exc())
            break
        continue
    return


__PLUGIN__ = "plugins.filters.main"
__help__ = "plugins.filters.help"
