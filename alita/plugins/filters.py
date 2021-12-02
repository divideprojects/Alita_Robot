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
from pyrogram.types import CallbackQuery, InlineKeyboardMarkup, Message

from alita.bot_class import LOGGER, Alita
from alita.database.filters_db import Filters
from alita.utils.cmd_senders import send_cmd
from alita.utils.custom_filters import admin_filter, command, owner_filter
from alita.utils.kbhelpers import ikb
from alita.utils.msg_types import Types, get_filter_type
from alita.utils.regex_utils import regex_searcher
from alita.utils.string import (build_keyboard,
                                escape_mentions_using_curly_brackets,
                                parse_button, split_quotes)

# Initialise
db = Filters()


@Alita.on_message(command("filters") & filters.group & ~filters.bot)
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
    return await m.reply_text(filters_chat, disable_web_page_preview=True)


@Alita.on_message(command(["filter", "addfilter"]) & admin_filter & ~filters.bot)
async def add_filter(_, m: Message):

    args = m.text.split(" ", 1)
    all_filters = db.get_all_filters(m.chat.id)
    actual_filters = {j for i in all_filters for j in i.split("|")}

    if (len(all_filters) >= 50) and (len(actual_filters) >= 150):
        await m.reply_text(
            "Only 50 filters and 150 aliases are allowed per chat!\nTo add more filters, remove the existing ones.",
        )
        return

    if not m.reply_to_message and len(m.text.split()) < 3:
        return await m.reply_text("Please read help section for how to save a filter!")

    if m.reply_to_message and len(args) < 2:
        return await m.reply_text("Please read help section for how to save a filter!")

    extracted = await split_quotes(args[1])
    keyword = extracted[0].lower()

    for k in keyword.split("|"):
        if k in actual_filters:
            return await m.reply_text(f"Filter <code>{k}</code> already exists!")

    if not keyword:
        return await m.reply_text(
            f"<code>{m.text}</code>\n\nError: You must give a name for this Filter!",
        )

    if keyword.startswith("<") or keyword.startswith(">"):
        return await m.reply_text("Cannot save a filter which starts with '<' or '>'")

    eee, msgtype, file_id = await get_filter_type(m)
    lol = eee if m.reply_to_message else extracted[1]
    teks = lol if msgtype == Types.TEXT else eee

    if not m.reply_to_message and msgtype == Types.TEXT and len(m.command) <= 2:
        return await m.reply_text(
            f"<code>{m.text}</code>\n\nError: There is no text in here!",
        )

    if not teks and not msgtype:
        return await m.reply_text(
            'Please provide keyword for this filter reply with!\nEnclose filter in <code>"double quotes"</code>',
        )

    if not msgtype:
        return await m.reply_text(
            "Please provide data for this filter reply with!",
        )

    add = db.save_filter(m.chat.id, keyword, teks, msgtype, file_id)
    LOGGER.info(f"{m.from_user.id} added new filter ({keyword}) in {m.chat.id}")
    if add:
        await m.reply_text(
            f"Saved filter for '<code>{', '.join(keyword.split('|'))}</code>' in <b>{m.chat.title}</b>!",
        )
    await m.stop_propagation()


@Alita.on_message(command(["stop", "unfilter"]) & admin_filter & ~filters.bot)
async def stop_filter(_, m: Message):
    args = m.command

    if len(args) < 1:
        return await m.reply_text("What should I stop replying to?")

    chat_filters = db.get_all_filters(m.chat.id)
    act_filters = {j for i in chat_filters for j in i.split("|")}

    if not chat_filters:
        return await m.reply_text("No filters active here!")

    for keyword in act_filters:
        if keyword == m.text.split(None, 1)[1].lower():
            db.rm_filter(m.chat.id, m.text.split(None, 1)[1].lower())
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
    command(
        ["rmallfilters", "removeallfilters", "stopall", "stopallfilters"],
    )
    & owner_filter,
)
async def rm_allfilters(_, m: Message):
    all_bls = db.get_all_filters(m.chat.id)
    if not all_bls:
        return await m.reply_text("No filters to stop in this chat.")

    return await m.reply_text(
        "Are you sure you want to clear all filters?",
        reply_markup=ikb(
            [[("⚠️ Confirm", "rm_allfilters"), ("❌ Cancel", "close_admin")]],
        ),
    )


@Alita.on_callback_query(filters.regex("^rm_allfilters$"))
async def rm_allfilters_callback(_, q: CallbackQuery):
    user_id = q.from_user.id
    user_status = (await q.message.chat.get_member(user_id)).status
    if user_status not in {"creator", "administrator"}:
        await q.answer(
            "You're not even an admin, don't try this explosive shit!",
            show_alert=True,
        )
        return
    if user_status != "creator":
        await q.answer(
            "You're just an admin, not owner\nStay in your limits!",
            show_alert=True,
        )
        return
    db.rm_all_filters(q.message.chat.id)
    await q.message.edit_text(f"Cleared all filters for {q.message.chat.title}")
    LOGGER.info(f"{user_id} removed all filter from {q.message.chat.id}")
    await q.answer("Cleared all Filters!", show_alert=True)
    return


async def send_filter_reply(c: Alita, m: Message, trigger: str):
    """Reply with assigned filter for the trigger"""
    getfilter = db.get_filter(m.chat.id, trigger)
    if m and not m.from_user:
        return

    if not getfilter:
        return await m.reply_text(
            "<b>Error:</b> Cannot find a type for this filter!!",
            quote=True,
        )

    msgtype = getfilter["msgtype"]
    if not msgtype:
        return await m.reply_text("<b>Error:</b> Cannot find a type for this filter!!")

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
        "id",
        "mention",
        "username",
        "chatname",
    ]
    text = await escape_mentions_using_curly_brackets(m, filter_reply, parse_words)
    teks, button = await parse_button(text)
    button = await build_keyboard(button)
    button = InlineKeyboardMarkup(button) if button else None
    textt = teks
    try:
        if msgtype == Types.TEXT:
            if button:
                try:
                    await m.reply_text(
                        textt,
                        # parse_mode="markdown",
                        reply_markup=button,
                        disable_web_page_preview=True,
                        quote=True,
                    )
                    return
                except RPCError as ef:
                    await m.reply_text(
                        "An error has occured! Cannot parse note.",
                        quote=True,
                    )
                    LOGGER.error(ef)
                    LOGGER.error(format_exc())
                    return
            else:
                await m.reply_text(
                    textt,
                    # parse_mode="markdown",
                    quote=True,
                    disable_web_page_preview=True,
                )
                return

        elif msgtype in (
            Types.STICKER,
            Types.VIDEO_NOTE,
            Types.CONTACT,
            Types.ANIMATED_STICKER,
        ):
            await (await send_cmd(c, msgtype))(
                m.chat.id,
                getfilter["fileid"],
                reply_markup=button,
                reply_to_message_id=m.message_id,
            )
        else:
            await (await send_cmd(c, msgtype))(
                m.chat.id,
                getfilter["fileid"],
                caption=textt,
                #   parse_mode="markdown",
                reply_markup=button,
                reply_to_message_id=m.message_id,
            )
    except Exception as ef:
        await m.reply_text(f"Error: {ef}")
        return msgtype

    return msgtype


@Alita.on_message(filters.text & filters.group & ~filters.bot, group=69)
async def filters_watcher(c: Alita, m: Message):

    chat_filters = db.get_all_filters(m.chat.id)
    actual_filters = {j for i in chat_filters for j in i.split("|")}

    for trigger in actual_filters:
        pattern = r"( |^|[^\w])" + re_escape(trigger) + r"( |$|[^\w])"
        match = await regex_searcher(pattern, m.text.lower())
        if match:
            try:
                msgtype = await send_filter_reply(c, m, trigger)
                LOGGER.info(f"Replied with {msgtype} to {trigger} in {m.chat.id}")
            except Exception as ef:
                await m.reply_text(f"Error: {ef}")
                LOGGER.error(ef)
                LOGGER.error(format_exc())
            break
        continue
    return


__PLUGIN__ = "filters"

_DISABLE_CMDS_ = ["filters"]

__alt_name__ = ["filters", "autoreply"]
