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
from alita.utils.custom_filters import admin_filter, owner_filter
from alita.utils.msg_types import Types, get_filter_type
from alita.utils.parser import mention_html
from alita.utils.regex_utils import regex_searcher
from alita.utils.string import parse_button, split_quotes

__PLUGIN__ = "plugins.filters.main"
__help__ = "plugins.filters.help"


# Initialise
db = Filters()


async def send_cmd(client: Alita, msgtype):
    GET_FORMAT = {
        Types.TEXT.value: client.send_message,
        Types.DOCUMENT.value: client.send_document,
        Types.PHOTO.value: client.send_photo,
        Types.VIDEO.value: client.send_video,
        Types.STICKER.value: client.send_sticker,
        Types.AUDIO.value: client.send_audio,
        Types.VOICE.value: client.send_voice,
        Types.VIDEO_NOTE.value: client.send_video_note,
        Types.ANIMATION.value: client.send_animation,
        Types.ANIMATED_STICKER.value: client.send_sticker,
        Types.CONTACT: client.send_contact,
    }
    return GET_FORMAT[msgtype]


@Alita.on_message(
    filters.command("filters", PREFIX_HANDLER) & filters.group & admin_filter,
)
async def view_blacklist(_, m: Message):

    filters_chat = f"Filters in {m.chat.title}"
    all_filters = db.get_all_filters(m.chat.id)

    if not all_filters:
        await m.reply_text(f"There are no filters in {m.chat.title}")
        return

    filters_chat += "\n".join(
        [f" • <code>{escape(i)}</code>" for i in all_filters],
    )

    await m.reply_text(filters_chat)
    return


@Alita.on_message(
    filters.command(["addfilter", "filter"], PREFIX_HANDLER)
    & filters.group
    & admin_filter,
)
async def add_filter(_, m: Message):

    args = m.text.split(None, 1)

    if not m.reply_to_message and len(args) < 2:
        await m.reply_text(
            "Please provide keyboard keyword for this filter to reply with!",
        )
        return

    if m.reply_to_message:
        if len(args) < 2:
            await m.reply_text(
                "Please provide keyword for this filter to reply with!",
            )
            return
        else:
            keyword = args[1]
    else:
        extracted = await split_quotes(args[1])
        keyword = extracted[0].lower()

    teks, msgtype, file_id = await get_filter_type(m)

    if not m.reply_to_message and len(args) >= 2:
        teks, _ = await parse_button(extracted[1])
        if not teks:
            await m.reply_text(
                "There is no filter message - You can't JUST have buttons, you need a message to go with it!",
            )
            return

    elif m.reply_to_message and len(args) >= 2:
        if m.reply_to_message.text:
            text_to_parsing = m.reply_to_message.text
        elif m.reply_to_message.caption:
            text_to_parsing = m.reply_to_message.caption
        else:
            text_to_parsing = ""
        teks, _ = await parse_button(text_to_parsing)

    elif not teks and not msgtype:
        await m.reply_text(
            "Please provide keyword for this filter reply with!",
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
    if add:
        await m.reply_text(
            f"Saved filter '{keyword}' in <b>{m.chat.title}</b>!",
        )
    await m.stop_propagation()


@Alita.on_message(
    filters.command(["rmfilter", "stop", "unfilter"], PREFIX_HANDLER)
    & filters.group
    & admin_filter,
)
async def stop_filter(_, m: Message):
    args = m.command

    if len(args) < 2:
        await m.reply_text("What should I stop replying to?")
        return

    chat_filters = db.get_all_filters(m.chat.id)

    if not chat_filters:
        await m.reply_text("No filters active here!")
        return

    for keyword in chat_filters:
        if keyword == args[1]:
            db.rm_filter(m.chat.id, args[1])
            await m.reply_text(
                f"Okay, I'll stop replying to that filter in <b>{m.chat.title}</b>.",
            )
            await m.stop_propagation()

    await m.reply_text(
        "That's not a filter - Click: /filters to get currently active filters.",
    )


@Alita.on_message(
    filters.command("rmallfilters", PREFIX_HANDLER) & filters.group & owner_filter,
)
async def rm_allfilters(_, m: Message):

    all_bls = db.get_all_filters(m.chat.id)
    if not all_bls:
        await m.reply_text("No notes are filters in this chat")
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
async def rm_allbl_callback(_, q: CallbackQuery):
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
    db.rm_all_blacklist(q.message.chat.id)
    await q.message.delete()
    await q.answer("Cleared all Filters!", show_alert=True)
    return


@Alita.on_message(filters.text & filters.group, group=11)
async def filters_watcher(c: Alita, m: Message):
    chat_filters = db.get_all_filters(m.chat.id)
    for trigger in chat_filters:
        pattern = r"( |^|[^\w])" + trigger + r"( |$|[^\w])"
        match = await regex_searcher(pattern, m.text.lower())
        if not match:
            continue
        try:
            await send_filter_reply(c, m, trigger)
        except RPCError as ef:
            LOGGER.error(ef)
            LOGGER.error(format_exc())
        break
    return


async def send_filter_reply(c: Alita, m: Message, trigger: str):
    """Reply with assigned filter for the trigger"""
    getnotes = db.get_filter(m.chat.id, trigger)
    msgtype = getnotes["msgtype"]
    if not getnotes:
        await m.reply_text(
            "<b>Error:</b> Cannot find a type for this note!!",
            quote=True,
        )
        return

    if msgtype == Types.TEXT:
        teks = getnotes["filter_reply"]
        await m.reply_text(teks, quote=True)
    elif msgtype in (
        Types.STICKER,
        Types.VIDEO_NOTE,
        Types.CONTACT,
        Types.ANIMATED_STICKER,
    ):
        await (await send_cmd(c, msgtype))(
            m.chat.id,
            getnotes["fileid"],
            reply_to_message_id=m.message_id,
        )
    else:
        if getnotes["filter_reply"]:
            teks = getnotes["filter_reply"]
        else:
            teks = ""
        await (await send_cmd(c, msgtype))(
            m.chat.id,
            getnotes["fileid"],
            caption=teks,
            parse_mode=None,
            reply_to_message_id=m.message_id,
        )
    return
